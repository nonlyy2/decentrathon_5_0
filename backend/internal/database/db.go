package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseURL string) (*pgxpool.Pool, error) {
	// pgx принимает только "postgres://", не "postgresql://"
	url := strings.Replace(databaseURL, "postgresql://", "postgres://", 1)

	// Railway требует SSL
	if !strings.Contains(url, "sslmode=") {
		if strings.Contains(url, "?") {
			url += "&sslmode=require"
		} else {
			url += "?sslmode=require"
		}
	}

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	config.MaxConns = 10

	// Railway стартует Postgres и бэкенд почти одновременно, первые попытки могут упасть
	const maxAttempts = 10
	const retryDelay = 3 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		pool, err := pgxpool.NewWithConfig(ctx, config)
		cancel()
		if err != nil {
			log.Printf("DB connect attempt %d/%d failed: %v — retrying in %s", attempt, maxAttempts, err, retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		pingErr := pool.Ping(ctx2)
		cancel2()
		if pingErr != nil {
			pool.Close()
			log.Printf("DB ping attempt %d/%d failed: %v — retrying in %s", attempt, maxAttempts, pingErr, retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("Database connected (attempt %d)", attempt)
		return pool, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxAttempts)
}
