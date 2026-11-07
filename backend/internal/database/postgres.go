package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/n1ckerr0r/chronograph-spb/backend/internal/config"
)

func New(cfg config.Config) (*pgxpool.Pool, error) {

	dsn := (&url.URL{
		Scheme: "postgres", User: url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host: fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort), Path: cfg.DBName,
		RawQuery: url.Values{"sslmode": []string{cfg.DBSSLMode}}.Encode(),
	}).String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			pool.Close()
		}
	}()

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
