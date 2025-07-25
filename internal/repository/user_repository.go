package repository

import (
	"context"
	"errors"
	"go-chat/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type pgxUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgxUserRepository{db: db}
}

func (r *pgxUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, email, password_hash)
		  	  VALUES ($1, $2, $3)
			  RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *pgxUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`

	user := new(domain.User)

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
	}

	return user, nil
}

var ErrUserNotFound = errors.New("user not found")