package sqlite3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"restaurant/pkg/migrator"
	"restaurant/services/auth/internal/model"
)

/* Интерфейс для транкзаций */
type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

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

/* Сохранить токен */
func (s *Storage) SaveRefreshToken(ctx context.Context, token *model.Token) error {
	return s.saveRefreshToken(ctx, s, token)
}

/* Получить токен(entity) по токену(string) */
func (s *Storage) GetRefreshToken(ctx context.Context, token string) (*model.Token, error) {
	query := `SELECT userid, token, revoked, expires_at, created_at
		FROM tokens
		WHERE token=?`

	var Token model.Token
	err := s.QueryRowContext(ctx, query, token).Scan(
		&Token.UserID, &Token.Token, &Token.Revoked, &Token.ExpiresAt, &Token.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &Token, nil
}

/* Отозвать токен */
func (s *Storage) RevokeRefreshToken(ctx context.Context, token string) error {
	return s.revokeRefreshToken(ctx, s, token)
}

/* Транкзация: Отозвать и сохранить токены */
func (s *Storage) RevokeAndSave(ctx context.Context, oldToken string, newToken *model.Token) error {
	tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start tx: %w", err)
	}
	defer tx.Rollback()

	err = s.revokeRefreshToken(ctx, tx, oldToken)
	if err != nil {
		return err
	}

	err = s.saveRefreshToken(ctx, tx, newToken)
	if err != nil {
		return err
	}

	return tx.Commit()
}

/* Сохранить токен */
func (s *Storage) saveRefreshToken(ctx context.Context, exec dbExecutor, token *model.Token) error {
	query := `INSERT INTO tokens
		(userid, token, revoked, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)`

	_, err := exec.ExecContext(ctx, query,
		token.UserID, token.Token, token.Revoked, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		if errors.Is(err, sqlite3.ErrConstraintUnique) {
			return model.TokenAlreadyExists
		}
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

/* Получить токен(entity) по токену(string) */
func (s *Storage) revokeRefreshToken(ctx context.Context, exec dbExecutor, token string) error {
	query := `UPDATE tokens SET
		revoked=1
		WHERE token=?`

	result, err := exec.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return model.ErrTokenNotFound
	}

	return nil
}
