package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already taken")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
}

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash, firstName string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func (repo *userRepository) CreateUser(ctx context.Context, email, passwordHash, firstName string) (string, error) {
	const insertUserQuery = `
		INSERT INTO users (email, password_hash, first_name)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var userID string
	err := repo.pool.QueryRow(ctx, insertUserQuery, strings.ToLower(email), passwordHash, firstName).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrEmailTaken
		}
		return "", err
	}
	return userID, nil
}

func (repo *userRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const getUserByEmailQuery = `
		SELECT id, email, password_hash, first_name
		FROM users
		WHERE lower(email) = lower($1)
	`

	var user User
	err := repo.pool.QueryRow(ctx, getUserByEmailQuery, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
