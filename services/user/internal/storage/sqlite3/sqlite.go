package sqlite3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"restaurant/pkg/migrator"
	"restaurant/services/user/internal/model"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

/* Структура хранилища */
type Storage struct {
	*sql.DB
}

/* Конструктор */
func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err := migrator.Migrate(db, migrator.SQLite3, migrationsFS, "migrations"); err != nil {
		return nil, err
	}

	return &Storage{db}, nil
}

/* Метод закрытия БД */
func (s *Storage) Close() error {
	return s.DB.Close()
}

/* Создание пользователя */
func (s *Storage) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users
		(id, name, password, created_at)
		VALUES (?, ?, ?, ?)`

	_, err := s.ExecContext(ctx, query, user.ID, user.Name, user.Password, user.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return model.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

/* Получение пользователя по name */
func (s *Storage) FindByName(ctx context.Context, name string) (*model.User, error) {
	query := `SELECT id, name, password, created_at
		FROM users
		WHERE name=?`

	var user model.User
	err := s.QueryRowContext(ctx, query, name).Scan(&user.ID, &user.Name, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
