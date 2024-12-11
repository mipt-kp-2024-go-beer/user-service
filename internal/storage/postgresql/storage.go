package database

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver import
	users "github.com/mipt-kp-2024-go-beer/user-service/internal"
	"github.com/mipt-kp-2024-go-beer/user-service/internal/oops"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dataSourceName string) (*Storage, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) LoadUsers(ctx context.Context) ([]users.User, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, login, password, permissions FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var output []users.User

	for rows.Next() {
		var user users.User
		if err := rows.Scan(&user.ID, &user.Login, &user.Password, &user.Permissions); err != nil {
			return nil, err
		}
		output = append(output, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, nil
}

func (s *Storage) CheckUser(ctx context.Context, user users.User) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1 AND password = $2", user.Login, user.Password).Scan(&id)

	if err == sql.ErrNoRows {
		return "", oops.ErrNoUser
	} else if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) CheckToken(ctx context.Context, access string) (users.Token, error) {
	var token users.Token
	bytes := []byte(access)
	err := s.db.QueryRowContext(ctx, "SELECT refresh_token, expiration FROM tokens WHERE access_token = $1", hex.EncodeToString(bytes)).Scan(&token.Refresh, &token.Expiration)

	fmt.Printf("%s", err)

	if err == sql.ErrNoRows {
		return users.Token{}, nil
	} else if err != nil {
		return users.Token{}, err
	}

	return token, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(ctx context.Context, user users.User) (id string, err error) {
	var existingID string
	err = s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", user.Login).Scan(&existingID)
	if err == nil {
		return existingID, oops.ErrDuplicateUser
	} else if err != sql.ErrNoRows {
		return "", err
	}

	err = s.db.QueryRowContext(ctx,
		"INSERT INTO users (login, password, permissions) VALUES ($1, $2, $3) RETURNING id",
		user.Login, user.Password, user.Permissions).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

func (s *Storage) LoadTokens(ctx context.Context) ([]users.Token, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT access_token, refresh_token, expiration FROM tokens")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []users.Token
	for rows.Next() {
		var token users.Token
		if err := rows.Scan(&token.Access, &token.Refresh, &token.Expiration); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Storage) GetSessionID(ctx context.Context, access string) (string, error) {
	var userID string

	err := s.db.QueryRowContext(ctx, "SELECT user_id FROM tokens WHERE access_token = $1", access).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", oops.ErrTokenExistance
	} else if err != nil {
		return "", err
	}

	return userID, nil
}

func (s *Storage) SaveToken(ctx context.Context, token users.Token, ID string) error {
	bytes_access := []byte(token.Access)
	bytes_refresh := []byte(token.Refresh)
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO tokens (access_token, refresh_token, expiration, user_id) VALUES ($1, $2, $3, $4)",
		hex.EncodeToString(bytes_access), hex.EncodeToString(bytes_refresh), token.Expiration, ID)

	fmt.Printf("%s", err)
	return err
}

func (s *Storage) TokenExpired(ctx context.Context, access string) (bool, error) {
	var expiration time.Time

	err := s.db.QueryRowContext(ctx, "SELECT expiration FROM tokens WHERE access_token = $1", access).Scan(&expiration)
	if err == sql.ErrNoRows {
		return false, oops.ErrTokenExistance
	} else if err != nil {
		return false, err
	}

	return time.Now().After(expiration), nil
}

func (s *Storage) PopToken(ctx context.Context, access string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM tokens WHERE access_token = $1", access)
	return err
}

func (s *Storage) User(ctx context.Context, ID string) (users.User, error) {
	var user users.User
	err := s.db.QueryRowContext(ctx, "SELECT id, login, password, permissions FROM users WHERE id = $1", ID).Scan(&user.ID, &user.Login, &user.Password, &user.Permissions)

	if err == sql.ErrNoRows {
		return users.User{}, oops.ErrNoUser
	} else if err != nil {
		return users.User{}, err
	}

	return user, nil
}

func (s *Storage) PopUser(ctx context.Context, ID string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return oops.ErrNoUser
	}

	return nil
}

func (s *Storage) ChangeUser(ctx context.Context, user users.User) (users.User, error) {
	_, err := s.db.ExecContext(ctx, "UPDATE users SET login = $1, password = $2, permissions = $3 WHERE id = $4",
		user.Login, user.Password, user.Permissions, user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, oops.ErrNoUser
		}
		return users.User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

func (s *Storage) SetPermission(ctx context.Context, ID string, Permissions uint) error {
	res, err := s.db.ExecContext(ctx, "UPDATE users SET permissions = $1 WHERE id = $2", Permissions, ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return oops.ErrNoUser
	}
	return nil
}
