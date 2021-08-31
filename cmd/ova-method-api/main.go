package main

import (
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"ova-method-api/internal"
	"ova-method-api/internal/app"
	igrpc "ova-method-api/pkg/ova-method-api"
)

var (
	server *grpc.Server
)

func main() {
	initLogger()

	cnf := getConfig()
	startGrpcServer(cnf)

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

func startGrpcServer(cnf *internal.Application) {
	listen, err := net.Listen("tcp", cnf.Grpc.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed create net listen")
	}

	server = grpc.NewServer()
	igrpc.RegisterOvaMethodApiServer(server, app.NewOvaMethodApi())

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
