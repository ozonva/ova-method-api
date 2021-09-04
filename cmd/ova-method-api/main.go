package main

import (
	"context"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	"ova-method-api/internal"
	"ova-method-api/internal/app"
	"ova-method-api/internal/app/middleware"
	"ova-method-api/internal/repo"
	igrpc "ova-method-api/pkg/ova-method-api"
)

var (
	conn          *sqlx.DB
	server        *grpc.Server
	tracingCloser io.Closer
)

func main() {
	cnf := getConfig()

	initLogger()
	initOpentracing(cnf)
	connectToDatabase(cnf)

	startGrpcServer(cnf, repo.NewMethodRepo(conn))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	defer signal.Stop(quit)
	<-quit

	shutdown()
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

func startGrpcServer(cnf *internal.Application, rep repo.MethodRepo) {
	listen, err := net.Listen("tcp", cnf.Grpc.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed create net listen")
	}

	tracing := middleware.NewTracingMiddleware(cnf.Tracing.GrpcEndpoints)
	server = grpc.NewServer(grpc.ChainUnaryInterceptor(tracing.UnaryIntercept))

	igrpc.RegisterOvaMethodApiServer(server, app.NewOvaMethodApi(rep))

	go func() {
		log.Info().Str("addr", cnf.Grpc.Addr).Msg("GRPC server started")
		if err = server.Serve(listen); err != nil {
			log.Fatal().Err(err).Msg("failed start grpc server")
		}
	}()
}

func shutdown() {
	server.GracefulStop()
	log.Info().Msg("GRPC server stopped")

	if err := conn.Close(); err != nil {
		log.Fatal().Err(err).Msg("failed close db connection")
	}

	if err := tracingCloser.Close(); err != nil {
		log.Fatal().Err(err).Msg("failed close opentracing")
	}
}
