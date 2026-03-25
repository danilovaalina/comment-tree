package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"comment-tree/internal/api"
	"comment-tree/internal/config"
	"comment-tree/internal/db/postgres"
	"comment-tree/internal/repository"
	"comment-tree/internal/service"

	"github.com/rs/zerolog/log"
)

func main() {
	cf, err := config.Load()
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Pool(ctx, cf.DatabaseURL)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}
	defer pool.Close()

	a := api.New(service.New(repository.New(pool), service.Options{
		DefaultLimit: cf.DefaultLimit,
		MaxLimit:     cf.MaxLimit,
	}))
	err = a.Start(cf.Addr)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}
}
