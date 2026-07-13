package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"restaurant/pkg/migrator"
	"restaurant/services/user/internal/model"

	"github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Storage struct {
	*sql.DB
}

type Config struct {
	Addr     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func New(cfg Config) (*Storage, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Addr, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	if err := migrator.Migrate(db, migrator.Postgres, migrationsFS, "migrations"); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return &Storage{db}, nil
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

func (s *Storage) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, name, password, created_at) 
              VALUES ($1, $2, $3, $4)`

	_, err := s.ExecContext(ctx, query, user.ID, user.Name, user.Password, user.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return model.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *Storage) FindByName(ctx context.Context, name string) (*model.User, error) {
	query := `SELECT id, name, password, created_at 
              FROM users 
              WHERE name = $1`

	var user model.User
	err := s.QueryRowContext(ctx, query, name).Scan(&user.ID, &user.Name, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}
