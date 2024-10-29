package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/romeulima/devbook-server/internal/api"
	"github.com/romeulima/devbook-server/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Error while execute code", "error", err)
		return
	}
	slog.Info("all systems off")
}

func run() error {

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"user=%s password=%s host=localhost port=5432 dbname=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	))

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	handler := api.NewHandler(storage.New(pool))

	s := http.Server{
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
		WriteTimeout: time.Second * 10,
		Addr:         ":8080",
		Handler:      handler,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
