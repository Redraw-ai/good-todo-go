package database

import (
	"context"
	"database/sql"
	"fmt"

	"good-todo-go/internal/ent"
	"good-todo-go/internal/infrastructure/environment"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
)

// DBClient holds both ent.Client and sql.DB for flexibility
type DBClient struct {
	Ent *ent.Client
	DB  *sql.DB
}

func NewEntClient(cfg *environment.Config) (*ent.Client, error) {
	dbClient, err := NewDBClient(cfg)
	if err != nil {
		return nil, err
	}
	return dbClient.Ent, nil
}

// NewDBClient creates both ent.Client and sql.DB from the same connection pool
func NewDBClient(cfg *environment.Config) (*DBClient, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Open sql.DB first
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	// Create ent client from the same sql.DB
	drv := entsql.OpenDB("postgres", db)
	client := ent.NewClient(ent.Driver(drv))

	// デバッグモードを有効化（開発時のみ）
	if cfg.AppEnv == "local" {
		client = client.Debug()
	}

	return &DBClient{
		Ent: client,
		DB:  db,
	}, nil
}

func CloseEntClient(client *ent.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}

func RunMigrations(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(ctx)
}
