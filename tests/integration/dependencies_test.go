//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestPostgresAndRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests require Docker")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	postgresContainer, err := postgres.Run(ctx, "postgres:18-alpine",
		postgres.WithDatabase("our_sell_test"),
		postgres.WithUsername("our_sell"),
		postgres.WithPassword("our_sell"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start PostgreSQL container: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(postgresContainer) })

	databaseURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get PostgreSQL connection string: %v", err)
	}
	connection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL container: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(context.Background()) })
	if err := connection.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL container: %v", err)
	}

	redisContainer, err := tcredis.Run(ctx, "redis:8-alpine")
	if err != nil {
		t.Fatalf("start Redis container: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(redisContainer) })
	redisURL, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get Redis connection string: %v", err)
	}
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse Redis connection string: %v", err)
	}
	client := redis.NewClient(options)
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping Redis container: %v", err)
	}
}
