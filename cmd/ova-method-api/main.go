package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Shopify/sarama"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

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
	cnf := getConfig()

	initLogger()
	initOpentracing(cnf)

	connectToQueue(cnf)
	connectToDatabase(cnf)

	startHttpServer(cnf)
	startGrpcServer(cnf, repo.NewMethodRepo(conn))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	defer signal.Stop(quit)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cnf.GetShutdownTime())
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

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

func initOpentracing(cnf *internal.Application) {
	cfg := &config.Configuration{
		ServiceName: cnf.Tracing.ServiceName,
		Disabled:    cnf.Tracing.Disabled,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		log.Fatal().Err(err).Msg("failed init tracer")
	}

	tracingCloser = closer
	opentracing.SetGlobalTracer(tracer)
}

func connectToQueue(cnf *internal.Application) {
	saramaCnf := sarama.NewConfig()
	saramaCnf.Producer.Return.Successes = true

	queue = iqueue.NewKafkaProvider(cnf.Kafka.Brokers, saramaCnf)

	if err := queue.Connect(); err != nil {
		log.Fatal().Err(err).Msg("failed connect to queue")
	}
}

func connectToDatabase(cnf *internal.Application) {
	ctx, cancel := context.WithTimeout(context.Background(), cnf.Database.GetConnTimeout())
	defer cancel()

	db, err := sqlx.Open(cnf.Database.Driver, cnf.Database.String())
	if err != nil {
		log.Fatal().Err(err).Msg("failed create db connection")
	}

	if err = db.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed ping db connection")
	}

	db.SetMaxOpenConns(cnf.Database.MaxOpenConns)
	db.SetMaxIdleConns(cnf.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cnf.Database.GetConnMaxLifetime())

	conn = db
}

func startHttpServer(cnf *internal.Application) {
	router := http.NewServeMux()
	router.Handle(cnf.Monitoring.HttpRoute, promhttp.Handler())

	httpServer = &http.Server{Addr: cnf.Http.Addr, Handler: router}

	go func() {
		log.Info().Str("addr", cnf.Http.Addr).Msg("HTTP server started")
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed start HTTP server")
		}
	}()
}

func startGrpcServer(cnf *internal.Application, rep repo.MethodRepo) {
	listen, err := net.Listen("tcp", cnf.Grpc.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed create net listen")
	}

	statusCounters := make([]monitoring.StatusCounter, 0, len(cnf.Monitoring.StatusCounters))
	for _, counter := range cnf.Monitoring.StatusCounters {
		statusCounters = append(statusCounters, monitoring.NewStatusCounter(
			counter.GrpcStatus,
			counter.GrpcEndpoints,
			promauto.NewCounter(prometheus.CounterOpts{
				Name: counter.Name,
				Help: counter.Desc,
			}),
		))
	}

	tracing := middleware.NewTracingMiddleware(cnf.Tracing.GrpcEndpoints)
	statusMonitoring := middleware.NewStatusMonitoringMiddleware(statusCounters)

	grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(tracing.UnaryIntercept, statusMonitoring.UnaryIntercept),
	)

	igrpc.RegisterOvaMethodApiServer(grpcServer, app.NewOvaMethodApi(rep, queue))

	go func() {
		log.Info().Str("addr", cnf.Grpc.Addr).Msg("GRPC server started")
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
