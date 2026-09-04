package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/vtech/our-sell/db/sqlc"
	"github.com/vtech/our-sell/internal/platform/apperror"
)

type Repository interface {
	Create(ctx context.Context, id uuid.UUID, email, passwordHash, name string, role Role) (User, error)
	ByEmail(ctx context.Context, email string) (User, error)
	ByID(ctx context.Context, id uuid.UUID) (User, error)
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
}

type SQLRepository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *SQLRepository { return &SQLRepository{queries: queries} }

func (r *SQLRepository) Create(ctx context.Context, id uuid.UUID, email, passwordHash, name string, role Role) (User, error) {
	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         string(role),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, apperror.ErrConflict
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return fromSQLUser(row), nil
}

func (r *SQLRepository) ByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, apperror.ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return fromSQLUser(row), nil
}

func (r *SQLRepository) ByID(ctx context.Context, id uuid.UUID) (User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, apperror.ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by ID: %w", err)
	}
	return fromSQLUser(row), nil
}

func (r *SQLRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.MarkUserEmailVerified(ctx, sqlc.MarkUserEmailVerifiedParams{ID: id, EmailVerifiedAt: pgtype.Timestamptz{Time: nowUTC(), Valid: true}}); err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	return nil
}

func (r *SQLRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	if err := r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{ID: id, PasswordHash: passwordHash}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func fromSQLUser(row sqlc.User) User {
	return User{
		ID:              row.ID.String(),
		Email:           row.Email,
		PasswordHash:    row.PasswordHash,
		Name:            row.Name,
		Role:            Role(row.Role),
		EmailVerifiedAt: nullableTime(row.EmailVerifiedAt),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func nowUTC() time.Time { return time.Now().UTC() }

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
