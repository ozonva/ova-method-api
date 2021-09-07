package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ova-method-api/internal"
	"ova-method-api/internal/app"
	"ova-method-api/internal/app/middleware"
	"ova-method-api/internal/monitoring"
	iqueue "ova-method-api/internal/queue"
	"ova-method-api/internal/repo"
	igrpc "ova-method-api/pkg/ova-method-api"
)

var (
	conn          *sqlx.DB
	tracingCloser io.Closer
	queue         iqueue.Queue
	httpServer    *http.Server
	grpcServer    *grpc.Server
)

func main() {
	config := getConfig()

	initLogger(config)
	initOpentracing(config)

	connectToQueue(config)
	connectToDatabase(config)

	startHttpServer(config)
	startGrpcServer(config, repo.NewMethodRepo(conn))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), config.GetShutdownTime())
	defer cancel()

	shutdown(ctx)
}

func getConfig() *internal.Application {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("failed load config")
	}
	return internal.LoadConfig(dir)
}

func initLogger(config *internal.Application) {
	var writer io.Writer

	switch config.Logging.Driver {
	case "file":
		if err := os.MkdirAll(config.Logging.FilePath, 0744); err != nil {
			log.Fatal().Err(err).Str("path", config.Logging.FilePath).Msg("failed create log directory")
		}

		writer = &lumberjack.Logger{
			MaxSize:    config.Logging.MaxSizeMb,
			MaxBackups: config.Logging.MaxBackups,
			MaxAge:     config.Logging.MaxAgeDay,
			Filename:   path.Join(config.Logging.FilePath, config.Logging.FileName),
		}
	default:
		writer = os.Stdout
	}

	level, err := zerolog.ParseLevel(config.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
		log.Error().Err(err).Str("level", config.Logging.Level).Msg("failed parse log level")
	}

	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: writer, TimeFormat: time.RFC3339})
}

func initOpentracing(config *internal.Application) {
	cfg := &jconfig.Configuration{
		ServiceName: config.Tracing.ServiceName,
		Disabled:    config.Tracing.Disabled,
		Sampler: &jconfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
	}

	tracer, closer, err := cfg.NewTracer(jconfig.Logger(jaeger.StdLogger))
	if err != nil {
		log.Fatal().Err(err).Msg("failed init tracer")
	}

	tracingCloser = closer
	opentracing.SetGlobalTracer(tracer)
}

func connectToQueue(config *internal.Application) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true

	queue = iqueue.NewKafkaProvider(config.Kafka.Brokers, saramaConfig)

	if err := queue.Connect(); err != nil {
		log.Fatal().Err(err).Msg("failed connect to queue")
	}
}

func connectToDatabase(config *internal.Application) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Database.GetConnTimeout())
	defer cancel()

	db, err := sqlx.Open(config.Database.Driver, config.Database.String())
	if err != nil {
		log.Fatal().Err(err).Msg("failed create db connection")
	}

	if err = db.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed ping db connection")
	}

	db.SetMaxOpenConns(config.Database.MaxOpenConns)
	db.SetMaxIdleConns(config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(config.Database.GetConnMaxLifetime())

	conn = db
}

func startHttpServer(config *internal.Application) {
	router := http.NewServeMux()
	router.Handle(config.Monitoring.HttpRoute, promhttp.Handler())

	httpServer = &http.Server{Addr: config.Http.Addr, Handler: router}

	go func() {
		log.Info().Str("addr", config.Http.Addr).Msg("HTTP server started")
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed start HTTP server")
		}
	}()
}

func startGrpcServer(config *internal.Application, rep repo.MethodRepo) {
	listen, err := net.Listen("tcp", config.Grpc.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed create net listen")
	}

	statusCounters := make([]monitoring.StatusCounter, 0, len(config.Monitoring.StatusCounters))
	for _, counter := range config.Monitoring.StatusCounters {
		statusCounters = append(statusCounters, monitoring.NewStatusCounter(
			counter.GrpcStatus,
			counter.GrpcEndpoints,
			promauto.NewCounter(prometheus.CounterOpts{
				Name: counter.Name,
				Help: counter.Desc,
			}),
		))
	}

	tracing := middleware.NewTracingMiddleware(config.Tracing.GrpcEndpoints)
	statusMonitoring := middleware.NewStatusMonitoringMiddleware(statusCounters)

	grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(tracing.UnaryIntercept, statusMonitoring.UnaryIntercept),
	)

	igrpc.RegisterOvaMethodApiServer(grpcServer, app.NewOvaMethodApi(rep, queue))

	go func() {
		log.Info().Str("addr", config.Grpc.Addr).Msg("GRPC server started")
		if err = grpcServer.Serve(listen); err != nil {
			log.Fatal().Err(err).Msg("failed start GRPC server")
		}
	}()
}

func shutdown(ctx context.Context) {
	grpcServer.GracefulStop()
	log.Info().Msg("GRPC server stopped")

	if err := conn.Close(); err != nil {
		log.Fatal().Err(err).Msg("failed close db connection")
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed shutdown http server")
	}
	log.Info().Msg("HTTP server stopped")

	if err := tracingCloser.Close(); err != nil {
		log.Fatal().Err(err).Msg("failed close opentracing")
	}

	if err := queue.Close(); err != nil {
		log.Fatal().Err(err).Msg("failed close connect to queue")
	}
}
