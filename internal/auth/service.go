package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vtech/our-sell/internal/email"
	"github.com/vtech/our-sell/internal/password"
	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/validator"
	"github.com/vtech/our-sell/internal/user"
)

type Service struct {
	users                 user.Repository
	sessions              Repository
	mailer                email.Mailer
	tokens                *TokenService
	verificationTokenTTL  time.Duration
	passwordResetTokenTTL time.Duration
	refreshTokenTTL       time.Duration
	now                   func() time.Time
}

func NewService(users user.Repository, sessions Repository, mailer email.Mailer, tokens *TokenService, verificationTokenTTL, passwordResetTokenTTL, refreshTokenTTL time.Duration) *Service {
	return &Service{
		users: users, sessions: sessions, mailer: mailer, tokens: tokens,
		verificationTokenTTL: verificationTokenTTL, passwordResetTokenTTL: passwordResetTokenTTL,
		refreshTokenTTL: refreshTokenTTL, now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (user.User, error) {
	emailAddress, name, err := validateRegistration(req)
	if err != nil {
		return user.User{}, err
	}
	hash, err := password.HashPassword(req.Password)
	if err != nil {
		return user.User{}, apperror.Wrap(err, "hash password")
	}
	userID, err := uuid.NewV7()
	if err != nil {
		return user.User{}, apperror.Wrap(err, "generate user ID")
	}
	created, err := s.users.Create(ctx, userID, emailAddress, hash, name, user.RoleUser)
	if err != nil {
		return user.User{}, err
	}
	verificationToken, err := randomToken()
	if err != nil {
		return user.User{}, apperror.Wrap(err, "generate verification token")
	}
	if err := s.sessions.ReplaceVerificationToken(ctx, userID, hashToken(verificationToken), s.clock().Add(s.verificationTokenTTL)); err != nil {
		return user.User{}, apperror.Wrap(err, "store verification token")
	}
	if err := s.mailer.SendVerification(ctx, created.Email, verificationToken); err != nil {
		return user.User{}, apperror.Wrap(err, "send verification email")
	}
	return created, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest, userAgent, ipAddress string) (ServiceResult, error) {
	emailAddress, err := validator.NormalizeEmail(req.Email)
	if err != nil || req.Password == "" || len(req.Password) > 128 {
		return ServiceResult{}, apperror.ErrInvalidCredentials
	}
	account, err := s.users.ByEmail(ctx, emailAddress)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			_, _ = password.VerifyPassword(req.Password, password.DummyHash())
			return ServiceResult{}, apperror.ErrInvalidCredentials
		}
		return ServiceResult{}, apperror.Wrap(err, "load account")
	}
	valid, err := password.VerifyPassword(req.Password, account.PasswordHash)
	if err != nil || !valid {
		return ServiceResult{}, apperror.ErrInvalidCredentials
	}
	if account.Role != user.RoleUser && account.Role != user.RoleAdmin {
		return ServiceResult{}, apperror.ErrUnauthorized
	}
	now := s.clock()
	userID, err := uuid.Parse(account.ID)
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "parse account ID")
	}
	sessionID, err := uuid.NewV7()
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "generate session ID")
	}
	accessToken, err := s.tokens.GenerateAccessToken(userID, sessionID, string(account.Role), now)
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "generate access token")
	}
	refreshToken, err := randomToken()
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "generate refresh token")
	}
	session := Session{
		ID: sessionID.String(), UserID: account.ID, UserAgent: truncate(userAgent, 512), IPAddress: truncate(ipAddress, 128),
		ExpiresAt: now.Add(s.refreshTokenTTL), LastUsedAt: now, CreatedAt: now,
	}
	if err := s.sessions.CreateSession(ctx, session, hashToken(refreshToken)); err != nil {
		return ServiceResult{}, apperror.Wrap(err, "create authentication session")
	}
	return ServiceResult{User: account, AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int64(s.tokens.ttl.Seconds())}, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (ServiceResult, error) {
	if len(rawRefreshToken) < 32 || len(rawRefreshToken) > 256 {
		return ServiceResult{}, apperror.ErrInvalidToken
	}
	now := s.clock()
	nextRefreshToken, err := randomToken()
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "generate refresh token")
	}
	session, reused, err := s.sessions.RotateRefreshToken(ctx, hashToken(rawRefreshToken), hashToken(nextRefreshToken), now)
	if err != nil {
		return ServiceResult{}, mapTokenError(err)
	}
	if reused {
		return ServiceResult{}, apperror.ErrInvalidToken
	}
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "parse session user ID")
	}
	sessionID, err := uuid.Parse(session.ID)
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "parse session ID")
	}
	account, err := s.users.ByID(ctx, userID)
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "load session user")
	}
	accessToken, err := s.tokens.GenerateAccessToken(userID, sessionID, string(account.Role), now)
	if err != nil {
		return ServiceResult{}, apperror.Wrap(err, "generate access token")
	}
	return ServiceResult{User: account, AccessToken: accessToken, RefreshToken: nextRefreshToken, ExpiresIn: int64(s.tokens.ttl.Seconds())}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	id, err := validator.UUID(sessionID)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	if err := s.sessions.RevokeSession(ctx, id); err != nil {
		return apperror.Wrap(err, "revoke session")
	}
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	id, err := validator.UUID(userID)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	if err := s.sessions.RevokeAllSessions(ctx, id); err != nil {
		return apperror.Wrap(err, "revoke all sessions")
	}
	return nil
}

func (s *Service) Sessions(ctx context.Context, userID string) ([]Session, error) {
	id, err := validator.UUID(userID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}
	result, err := s.sessions.ListActiveSessions(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, "list sessions")
	}
	return result, nil
}

func (s *Service) RevokeUserSession(ctx context.Context, userID, sessionID string) error {
	uid, err := validator.UUID(userID)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	sid, err := validator.UUID(sessionID)
	if err != nil {
		return apperror.ErrNotFound
	}
	if err := s.sessions.RevokeUserSession(ctx, uid, sid); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return apperror.Wrap(err, "revoke user session")
	}
	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	if !validTokenInput(rawToken) {
		return apperror.ErrInvalidToken
	}
	if _, err := s.sessions.VerifyEmail(ctx, hashToken(rawToken), s.clock()); err != nil {
		return mapTokenError(err)
	}
	return nil
}

func (s *Service) ResendVerification(ctx context.Context, rawEmail string) error {
	emailAddress, err := validator.NormalizeEmail(rawEmail)
	if err != nil {
		return nil
	}
	account, err := s.users.ByEmail(ctx, emailAddress)
	if errors.Is(err, apperror.ErrNotFound) {
		return nil
	}
	if err != nil {
		return apperror.Wrap(err, "load verification account")
	}
	if account.EmailVerifiedAt != nil {
		return nil
	}
	rawToken, err := randomToken()
	if err != nil {
		return apperror.Wrap(err, "generate verification token")
	}
	userID, err := uuid.Parse(account.ID)
	if err != nil {
		return apperror.Wrap(err, "parse verification user ID")
	}
	if err := s.sessions.ReplaceVerificationToken(ctx, userID, hashToken(rawToken), s.clock().Add(s.verificationTokenTTL)); err != nil {
		return apperror.Wrap(err, "store verification token")
	}
	if err := s.mailer.SendVerification(ctx, account.Email, rawToken); err != nil {
		return apperror.Wrap(err, "send verification email")
	}
	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, rawEmail string) error {
	emailAddress, err := validator.NormalizeEmail(rawEmail)
	if err != nil {
		return nil
	}
	account, err := s.users.ByEmail(ctx, emailAddress)
	if errors.Is(err, apperror.ErrNotFound) {
		return nil
	}
	if err != nil {
		return apperror.Wrap(err, "load reset account")
	}
	rawToken, err := randomToken()
	if err != nil {
		return apperror.Wrap(err, "generate reset token")
	}
	userID, err := uuid.Parse(account.ID)
	if err != nil {
		return apperror.Wrap(err, "parse reset user ID")
	}
	if err := s.sessions.ReplacePasswordResetToken(ctx, userID, hashToken(rawToken), s.clock().Add(s.passwordResetTokenTTL)); err != nil {
		return apperror.Wrap(err, "store reset token")
	}
	if err := s.mailer.SendPasswordReset(ctx, account.Email, rawToken); err != nil {
		return apperror.Wrap(err, "send reset email")
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if !validTokenInput(req.Token) {
		return apperror.ErrInvalidToken
	}
	if err := validator.Password(req.NewPassword); err != nil {
		return apperror.NewValidation(map[string]string{"new_password": err.Error()})
	}
	hash, err := password.HashPassword(req.NewPassword)
	if err != nil {
		return apperror.Wrap(err, "hash reset password")
	}
	if _, err := s.sessions.ResetPassword(ctx, hashToken(req.Token), hash, s.clock()); err != nil {
		return mapTokenError(err)
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID, sessionID string, req ChangePasswordRequest) error {
	uid, err := validator.UUID(userID)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	sid, err := validator.UUID(sessionID)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	if err := validator.Password(req.NewPassword); err != nil {
		return apperror.NewValidation(map[string]string{"new_password": err.Error()})
	}
	account, err := s.users.ByID(ctx, uid)
	if err != nil {
		return apperror.Wrap(err, "load current account")
	}
	valid, err := password.VerifyPassword(req.CurrentPassword, account.PasswordHash)
	if err != nil || !valid {
		return apperror.ErrInvalidCredentials
	}
	hash, err := password.HashPassword(req.NewPassword)
	if err != nil {
		return apperror.Wrap(err, "hash new password")
	}
	if err := s.sessions.ChangePassword(ctx, uid, sid, hash); err != nil {
		return apperror.Wrap(err, "change password")
	}
	return nil
}

func (s *Service) Me(ctx context.Context, userID string) (user.User, error) {
	id, err := validator.UUID(userID)
	if err != nil {
		return user.User{}, apperror.ErrUnauthorized
	}
	account, err := s.users.ByID(ctx, id)
	if err != nil {
		return user.User{}, apperror.Wrap(err, "load current account")
	}
	return account, nil
}

func (s *Service) clock() time.Time { return s.now().UTC() }

func validateRegistration(req RegisterRequest) (string, string, error) {
	emailAddress, err := validator.NormalizeEmail(req.Email)
	if err != nil {
		return "", "", apperror.NewValidation(map[string]string{"email": err.Error()})
	}
	if err := validator.Password(req.Password); err != nil {
		return "", "", apperror.NewValidation(map[string]string{"password": err.Error()})
	}
	name, err := validator.Name(req.Name)
	if err != nil {
		return "", "", apperror.NewValidation(map[string]string{"name": err.Error()})
	}
	return emailAddress, name, nil
}

func randomToken() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("read secure random bytes: %w", err)
	}
	return fmt.Sprintf("%x", data), nil
}

func hashToken(raw string) []byte {
	digest := sha256.Sum256([]byte(raw))
	return digest[:]
}

func validTokenInput(value string) bool { return len(value) >= 32 && len(value) <= 256 }

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func mapTokenError(err error) error {
	if errors.Is(err, apperror.ErrInvalidToken) {
		return apperror.ErrInvalidToken
	}
	return apperror.Wrap(err, "authentication token operation")
}
