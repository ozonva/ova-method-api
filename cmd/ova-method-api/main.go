package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"ova-method-api/internal/model"
	"ova-method-api/internal/repo"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"ova-method-api/internal"
	"ova-method-api/internal/app"
	igrpc "ova-method-api/pkg/ova-method-api"
)

var (
	conn   *sqlx.DB
	server *grpc.Server
)

func main() {
	initLogger()

	cnf := getConfig()
	connectToDatabase(cnf)

	startGrpcServer(cnf, repo.NewMethodRepo(conn))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	defer signal.Stop(quit)
	<-quit

	shutdown()
}

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

func getConfig() *internal.Application {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("failed load config")
	}
	return internal.LoadConfig(dir)
}

func connectToDatabase(cnf *internal.Application) {
	db, err := sqlx.Connect(cnf.Database.Driver, cnf.Database.String())
	if err != nil {
		log.Fatal().Err(err).Msg("failed create db connect")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed ping db connection")
	}

	db.SetMaxOpenConns(cnf.Database.MaxOpenConns)
	db.SetMaxIdleConns(cnf.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cnf.Database.ConnMaxLifetime) * time.Second)

	conn = db
}

func startGrpcServer(cnf *internal.Application, rep repo.MethodRepo) {
	err := rep.Add([]model.Method{{UserId: 1, Value: "1"}, {UserId: 1, Value: "2"}})
	if err != nil {
		log.Fatal().Err(err).Msg("insert")
	}

	listen, err := net.Listen("tcp", cnf.Grpc.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed create net listen")
	}

	server = grpc.NewServer()
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
}
