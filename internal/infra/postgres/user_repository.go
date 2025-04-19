package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"time"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) InsertUser(ctx context.Context, user *domain.User) (string, error) {
	query := `
		INSERT INTO users (name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var id string
	err := r.db.QueryRow(ctx, query,
		user.Name,
		user.Email,
		user.Password,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = $1`
	user := &domain.User{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Created,
		&user.Updated,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`
	user := &domain.User{}

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Created,
		&user.Updated,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password = $3, updated_at = $4
		WHERE id = $5`

	_, err := r.db.Exec(ctx, query,
		user.Name,
		user.Email,
		user.Password,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return err
	}

	return nil
}
