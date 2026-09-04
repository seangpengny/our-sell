package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vtech/our-sell/internal/email"
	"github.com/vtech/our-sell/internal/password"
	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/user"
)

type fakeUsers struct {
	accounts map[string]user.User
}

func (f *fakeUsers) Create(_ context.Context, id uuid.UUID, emailAddress, passwordHash, name string, role user.Role) (user.User, error) {
	if _, exists := f.accounts[emailAddress]; exists {
		return user.User{}, apperror.ErrConflict
	}
	account := user.User{ID: id.String(), Email: emailAddress, PasswordHash: passwordHash, Name: name, Role: role, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	f.accounts[emailAddress] = account
	return account, nil
}

func (f *fakeUsers) ByEmail(_ context.Context, emailAddress string) (user.User, error) {
	account, ok := f.accounts[emailAddress]
	if !ok {
		return user.User{}, apperror.ErrNotFound
	}
	return account, nil
}

func (f *fakeUsers) ByID(_ context.Context, id uuid.UUID) (user.User, error) {
	for _, account := range f.accounts {
		if account.ID == id.String() {
			return account, nil
		}
	}
	return user.User{}, apperror.ErrNotFound
}

func (f *fakeUsers) MarkEmailVerified(context.Context, uuid.UUID) error { return nil }

func (f *fakeUsers) UpdatePassword(_ context.Context, id uuid.UUID, hash string) error {
	for emailAddress, account := range f.accounts {
		if account.ID == id.String() {
			account.PasswordHash = hash
			f.accounts[emailAddress] = account
			return nil
		}
	}
	return apperror.ErrNotFound
}

type fakeSession struct {
	session Session
	used    bool
}

type fakeSessions struct {
	byToken map[string]*fakeSession
}

func (f *fakeSessions) CreateSession(_ context.Context, session Session, tokenHash []byte) error {
	f.byToken[string(tokenHash)] = &fakeSession{session: session}
	return nil
}

func (f *fakeSessions) RotateRefreshToken(_ context.Context, presentedHash, nextHash []byte, now time.Time) (Session, bool, error) {
	current, ok := f.byToken[string(presentedHash)]
	if !ok {
		return Session{}, false, apperror.ErrInvalidToken
	}
	if current.used {
		return current.session, true, nil
	}
	if !current.session.ExpiresAt.After(now) {
		return Session{}, false, apperror.ErrInvalidToken
	}
	current.used = true
	next := &fakeSession{session: current.session}
	next.session.LastUsedAt = now
	f.byToken[string(nextHash)] = next
	return next.session, false, nil
}

func (f *fakeSessions) RevokeSession(context.Context, uuid.UUID) error                  { return nil }
func (f *fakeSessions) RevokeAllSessions(context.Context, uuid.UUID) error              { return nil }
func (f *fakeSessions) RevokeOtherSessions(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (f *fakeSessions) ChangePassword(context.Context, uuid.UUID, uuid.UUID, string) error {
	return nil
}
func (f *fakeSessions) ListActiveSessions(context.Context, uuid.UUID) ([]Session, error) {
	return nil, nil
}
func (f *fakeSessions) RevokeUserSession(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (f *fakeSessions) ReplaceVerificationToken(context.Context, uuid.UUID, []byte, time.Time) error {
	return nil
}
func (f *fakeSessions) VerifyEmail(context.Context, []byte, time.Time) (uuid.UUID, error) {
	return uuid.Nil, apperror.ErrInvalidToken
}
func (f *fakeSessions) ReplacePasswordResetToken(context.Context, uuid.UUID, []byte, time.Time) error {
	return nil
}
func (f *fakeSessions) ResetPassword(context.Context, []byte, string, time.Time) (uuid.UUID, error) {
	return uuid.Nil, apperror.ErrInvalidToken
}

type fakeMailer struct{}

func (fakeMailer) SendVerification(context.Context, string, string) error  { return nil }
func (fakeMailer) SendPasswordReset(context.Context, string, string) error { return nil }

var _ email.Mailer = fakeMailer{}

func newTestService(t *testing.T) (*Service, *fakeSessions) {
	t.Helper()
	users := &fakeUsers{accounts: make(map[string]user.User)}
	sessions := &fakeSessions{byToken: make(map[string]*fakeSession)}
	tokens, err := NewTokenService("01234567890123456789012345678901", "our-sell-api", "our-sell-client", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(users, sessions, fakeMailer{}, tokens, time.Hour, time.Hour, 24*time.Hour)
	return service, sessions
}

func TestLoginUsesGenericErrorForUnknownAndWrongCredentials(t *testing.T) {
	service, _ := newTestService(t)
	hash, err := password.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	service.users.(*fakeUsers).accounts["user@example.com"] = user.User{ID: uuid.New().String(), Email: "user@example.com", PasswordHash: hash, Name: "Test User", Role: user.RoleUser}

	for name, request := range map[string]LoginRequest{
		"wrong password":  {Email: "user@example.com", Password: "wrong password"},
		"unknown account": {Email: "nobody@example.com", Password: "correct horse battery staple"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.Login(context.Background(), request, "test", "127.0.0.1")
			if !errors.Is(err, apperror.ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v", err)
			}
		})
	}
}

func TestRefreshRotatesAndRejectsReuse(t *testing.T) {
	service, sessions := newTestService(t)
	account := user.User{ID: uuid.New().String(), Email: "user@example.com", PasswordHash: password.DummyHash(), Name: "Test User", Role: user.RoleUser}
	service.users.(*fakeUsers).accounts[account.Email] = account
	result, err := service.Login(context.Background(), LoginRequest{Email: account.Email, Password: "invalid-login-password"}, "test", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	next, err := service.Refresh(context.Background(), result.RefreshToken)
	if err != nil || next.RefreshToken == result.RefreshToken {
		t.Fatalf("Refresh() = %+v, %v", next, err)
	}
	if _, err := service.Refresh(context.Background(), result.RefreshToken); !errors.Is(err, apperror.ErrInvalidToken) {
		t.Fatalf("reused refresh token error = %v", err)
	}
	if len(sessions.byToken) != 2 {
		t.Fatalf("stored refresh token history = %d, want 2", len(sessions.byToken))
	}
}
