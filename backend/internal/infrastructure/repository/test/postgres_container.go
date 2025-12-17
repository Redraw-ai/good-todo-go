package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
)

type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DB        *sql.DB
	DSN       string
}

func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	dbName := "test_db"
	dbUser := "test_user"
	dbPassword := "test_password"

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresContainer{
		Container: container,
		DB:        db,
		DSN:       connStr,
	}, nil
}

func (pc *PostgresContainer) Close(ctx context.Context) error {
	if pc.DB != nil {
		pc.DB.Close()
	}
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}

// getMigrationsDir returns the path to the migrations directory
func getMigrationsDir() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	// Navigate from internal/infrastructure/repository/test to internal/ent/migrate/migrations
	testDir := filepath.Dir(currentFile)
	backendDir := filepath.Join(testDir, "..", "..", "..", "..")
	migrationsDir := filepath.Join(backendDir, "internal", "ent", "migrate", "migrations")

	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// RunMigrations applies Atlas migration files in order
func (pc *PostgresContainer) RunMigrations(ctx context.Context) error {
	migrationsDir, err := getMigrationsDir()
	if err != nil {
		return fmt.Errorf("failed to get migrations directory: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Get all .sql files and sort them
	var sqlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlFiles = append(sqlFiles, entry.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Execute each migration file
	for _, fileName := range sqlFiles {
		filePath := filepath.Join(migrationsDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		// Skip empty files
		sql := strings.TrimSpace(string(content))
		if sql == "" {
			continue
		}

		// Handle database-specific commands that might fail in test environment
		// Skip GRANT CONNECT since test_db user is the owner
		sql = strings.ReplaceAll(sql, "GRANT CONNECT ON DATABASE goodtodo_dev TO goodtodo_app;", "")

		if _, err := pc.DB.ExecContext(ctx, sql); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}
	}

	return nil
}

// GetAppUserDSN returns DSN for the app user (RLS enforced)
func (pc *PostgresContainer) GetAppUserDSN(ctx context.Context) (string, error) {
	host, err := pc.Container.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := pc.Container.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgres://goodtodo_app:app_secret@%s:%s/test_db?sslmode=disable", host, port.Port()), nil
}
