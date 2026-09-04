package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vtech/our-sell/db/sqlc"
	"github.com/vtech/our-sell/internal/platform/apperror"
)

type Repository interface {
	CreateSession(ctx context.Context, session Session, tokenHash []byte) error
	RotateRefreshToken(ctx context.Context, presentedHash, nextHash []byte, now time.Time) (Session, bool, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
	RevokeOtherSessions(ctx context.Context, userID, currentSessionID uuid.UUID) error
	ChangePassword(ctx context.Context, userID, currentSessionID uuid.UUID, passwordHash string) error
	ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]Session, error)
	RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error
	ReplaceVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	VerifyEmail(ctx context.Context, tokenHash []byte, now time.Time) (uuid.UUID, error)
	ReplacePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	ResetPassword(ctx context.Context, tokenHash []byte, passwordHash string, now time.Time) (uuid.UUID, error)
}

type SQLRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *sqlc.Queries, pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{queries: queries, pool: pool}
}

func (r *SQLRepository) CreateSession(ctx context.Context, session Session, tokenHash []byte) error {
	sessionID, err := uuid.Parse(session.ID)
	if err != nil {
		return fmt.Errorf("parse session ID: %w", err)
	}
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		return fmt.Errorf("parse session user ID: %w", err)
	}
	return r.inTx(ctx, func(q *sqlc.Queries) error {
		if err := q.CreateSession(ctx, sqlc.CreateSessionParams{
			ID: sessionID, UserID: userID, TokenHash: tokenHash,
			UserAgent: session.UserAgent, IpAddress: session.IPAddress,
			ExpiresAt: timestamptz(session.ExpiresAt), LastUsedAt: timestamptz(session.LastUsedAt),
		}); err != nil {
			return fmt.Errorf("insert session: %w", err)
		}
		if err := q.CreateSessionToken(ctx, sqlc.CreateSessionTokenParams{
			ID: uuid.New(), SessionID: sessionID, TokenHash: tokenHash,
		}); err != nil {
			return fmt.Errorf("insert session token: %w", err)
		}
		return nil
	})
}

func (r *SQLRepository) RotateRefreshToken(ctx context.Context, presentedHash, nextHash []byte, now time.Time) (Session, bool, error) {
	var result Session
	var reused bool
	err := r.inTx(ctx, func(q *sqlc.Queries) error {
		row, err := q.GetSessionToken(ctx, presentedHash)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrInvalidToken
		}
		if err != nil {
			return fmt.Errorf("find refresh session: %w", err)
		}
		result = sessionFromTokenRow(row)
		if row.TokenUsedAt.Valid || row.TokenRevokedAt.Valid {
			reused = true
			if err := q.RevokeAllSessions(ctx, row.UserID); err != nil {
				return fmt.Errorf("revoke sessions after refresh-token reuse: %w", err)
			}
			return nil
		}
		if row.RevokedAt.Valid || !row.ExpiresAt.Valid || !row.ExpiresAt.Time.After(now) {
			return apperror.ErrInvalidToken
		}
		if err := q.MarkSessionTokenUsed(ctx, sqlc.MarkSessionTokenUsedParams{
			SessionID: row.ID, TokenHash: presentedHash, UsedAt: timestamptz(now),
		}); err != nil {
			return fmt.Errorf("mark refresh token used: %w", err)
		}
		if err := q.UpdateSessionToken(ctx, sqlc.UpdateSessionTokenParams{
			ID: row.ID, TokenHash: nextHash, LastUsedAt: timestamptz(now),
		}); err != nil {
			return fmt.Errorf("rotate session token: %w", err)
		}
		if err := q.CreateSessionToken(ctx, sqlc.CreateSessionTokenParams{
			ID: uuid.New(), SessionID: row.ID, TokenHash: nextHash,
		}); err != nil {
			return fmt.Errorf("store rotated session token: %w", err)
		}
		result.LastUsedAt = now
		return nil
	})
	return result, reused, err
}

func (r *SQLRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := r.queries.RevokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *SQLRepository) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	if err := r.queries.RevokeAllSessions(ctx, userID); err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}
	return nil
}

func (r *SQLRepository) RevokeOtherSessions(ctx context.Context, userID, currentSessionID uuid.UUID) error {
	if err := r.queries.RevokeOtherSessions(ctx, sqlc.RevokeOtherSessionsParams{UserID: userID, ID: currentSessionID}); err != nil {
		return fmt.Errorf("revoke other sessions: %w", err)
	}
	return nil
}

func (r *SQLRepository) ChangePassword(ctx context.Context, userID, currentSessionID uuid.UUID, passwordHash string) error {
	return r.inTx(ctx, func(q *sqlc.Queries) error {
		if err := q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{ID: userID, PasswordHash: passwordHash}); err != nil {
			return fmt.Errorf("update password: %w", err)
		}
		if err := q.RevokeOtherSessions(ctx, sqlc.RevokeOtherSessionsParams{UserID: userID, ID: currentSessionID}); err != nil {
			return fmt.Errorf("revoke other sessions: %w", err)
		}
		return nil
	})
}

func (r *SQLRepository) ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	rows, err := r.queries.ListActiveSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list active sessions: %w", err)
	}
	result := make([]Session, 0, len(rows))
	for _, row := range rows {
		result = append(result, sessionFromRow(row))
	}
	return result, nil
}

func (r *SQLRepository) RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	commandTag, err := r.queries.RevokeUserSession(ctx, sqlc.RevokeUserSessionParams{ID: sessionID, UserID: userID})
	if err != nil {
		return fmt.Errorf("revoke user session: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) ReplaceVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	return r.inTx(ctx, func(q *sqlc.Queries) error {
		if err := q.InvalidateVerificationTokens(ctx, userID); err != nil {
			return fmt.Errorf("invalidate verification tokens: %w", err)
		}
		if err := q.CreateVerificationToken(ctx, sqlc.CreateVerificationTokenParams{
			ID: uuid.New(), UserID: userID, TokenHash: tokenHash, ExpiresAt: timestamptz(expiresAt),
		}); err != nil {
			return fmt.Errorf("create verification token: %w", err)
		}
		return nil
	})
}

func (r *SQLRepository) VerifyEmail(ctx context.Context, tokenHash []byte, now time.Time) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.inTx(ctx, func(q *sqlc.Queries) error {
		row, err := q.GetVerificationToken(ctx, tokenHash)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrInvalidToken
		}
		if err != nil {
			return fmt.Errorf("find verification token: %w", err)
		}
		if row.UsedAt.Valid || !row.ExpiresAt.Valid || !row.ExpiresAt.Time.After(now) {
			return apperror.ErrInvalidToken
		}
		if err := q.MarkVerificationTokenUsed(ctx, sqlc.MarkVerificationTokenUsedParams{ID: row.ID, UsedAt: timestamptz(now)}); err != nil {
			return fmt.Errorf("mark verification token used: %w", err)
		}
		if err := q.MarkUserEmailVerified(ctx, sqlc.MarkUserEmailVerifiedParams{ID: row.UserID, EmailVerifiedAt: timestamptz(now)}); err != nil {
			return fmt.Errorf("mark user verified: %w", err)
		}
		userID = row.UserID
		return nil
	})
	return userID, err
}

func (r *SQLRepository) ReplacePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	return r.inTx(ctx, func(q *sqlc.Queries) error {
		if err := q.InvalidatePasswordResetTokens(ctx, userID); err != nil {
			return fmt.Errorf("invalidate reset tokens: %w", err)
		}
		if err := q.CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
			ID: uuid.New(), UserID: userID, TokenHash: tokenHash, ExpiresAt: timestamptz(expiresAt),
		}); err != nil {
			return fmt.Errorf("create password reset token: %w", err)
		}
		return nil
	})
}

func (r *SQLRepository) ResetPassword(ctx context.Context, tokenHash []byte, passwordHash string, now time.Time) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.inTx(ctx, func(q *sqlc.Queries) error {
		row, err := q.GetPasswordResetToken(ctx, tokenHash)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrInvalidToken
		}
		if err != nil {
			return fmt.Errorf("find password reset token: %w", err)
		}
		if row.UsedAt.Valid || !row.ExpiresAt.Valid || !row.ExpiresAt.Time.After(now) {
			return apperror.ErrInvalidToken
		}
		if err := q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{ID: row.UserID, PasswordHash: passwordHash}); err != nil {
			return fmt.Errorf("update password in reset: %w", err)
		}
		if err := q.MarkPasswordResetTokenUsed(ctx, sqlc.MarkPasswordResetTokenUsedParams{ID: row.ID, UsedAt: timestamptz(now)}); err != nil {
			return fmt.Errorf("mark reset token used: %w", err)
		}
		if err := q.RevokeAllSessions(ctx, row.UserID); err != nil {
			return fmt.Errorf("revoke sessions after reset: %w", err)
		}
		userID = row.UserID
		return nil
	})
	return userID, err
}

func (r *SQLRepository) inTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	queries := r.queries.WithTx(tx)
	if err := fn(queries); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func sessionFromRow(row sqlc.ListActiveSessionsRow) Session {
	return Session{ID: row.ID.String(), UserID: row.UserID.String(), UserAgent: row.UserAgent, IPAddress: row.IpAddress, ExpiresAt: row.ExpiresAt.Time, LastUsedAt: row.LastUsedAt.Time, CreatedAt: row.CreatedAt.Time, RevokedAt: nullableTime(row.RevokedAt)}
}

func sessionFromTokenRow(row sqlc.GetSessionTokenRow) Session {
	return Session{ID: row.ID.String(), UserID: row.UserID.String(), UserAgent: row.UserAgent, IPAddress: row.IpAddress, ExpiresAt: row.ExpiresAt.Time, LastUsedAt: row.LastUsedAt.Time, CreatedAt: row.CreatedAt.Time, RevokedAt: nullableTime(row.RevokedAt)}
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
