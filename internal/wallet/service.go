package wallet

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vtech/our-sell/internal/config"
	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/validator"
)

type Service struct {
	repository Repository
	provider   Provider
	config     config.Config
	logger     *slog.Logger
	now        func() time.Time
}

func NewService(repository Repository, provider Provider, cfg config.Config, logger *slog.Logger) *Service {
	return &Service{repository: repository, provider: provider, config: cfg, logger: logger, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Wallet(ctx context.Context, userID string) (Wallet, error) {
	uid, err := validator.UUID(userID)
	if err != nil {
		return Wallet{}, apperror.ErrUnauthorized
	}
	wallet, err := s.repository.EnsureWallet(ctx, uid)
	if err != nil {
		return Wallet{}, apperror.Wrap(err, "load wallet")
	}
	return wallet, nil
}

func (s *Service) Ledger(ctx context.Context, userID string, page, pageSize int) (LedgerResult, error) {
	uid, err := validator.UUID(userID)
	if err != nil {
		return LedgerResult{}, apperror.ErrUnauthorized
	}
	page, pageSize = normalizePagination(page, pageSize)
	entries, hasMore, err := s.repository.ListLedger(ctx, uid, page, pageSize)
	if err != nil {
		return LedgerResult{}, apperror.Wrap(err, "load wallet ledger")
	}
	return LedgerResult{Entries: entries, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}

func (s *Service) CreateTopup(ctx context.Context, userID string, input TopupInput) (Topup, error) {
	uid, err := validator.UUID(userID)
	if err != nil {
		return Topup{}, apperror.ErrUnauthorized
	}
	if s.provider == nil {
		return Topup{}, apperror.ErrDependencyUnavailable
	}
	if err := s.validateAmount(input.AmountUSD); err != nil {
		return Topup{}, err
	}
	now := s.now()
	expiresAt := now.Add(s.config.BakongTopupTTL)
	id, err := uuid.NewV7()
	if err != nil {
		return Topup{}, apperror.Wrap(err, "generate wallet top-up ID")
	}
	// KHQR Bill Number is limited to 25 characters. UUIDv7 still gives us
	// ample uniqueness when represented by the first 22 hexadecimal characters.
	reference := "WT-" + strings.ToUpper(strings.ReplaceAll(id.String(), "-", ""))[:22]
	qrPayload, qrMD5, err := s.provider.GenerateQR(reference, input.AmountUSD, expiresAt)
	if err != nil {
		return Topup{}, apperror.Wrap(err, "generate Bakong KHQR")
	}
	wallet, err := s.repository.EnsureWallet(ctx, uid)
	if err != nil {
		return Topup{}, apperror.Wrap(err, "prepare wallet for top-up")
	}
	topup, err := s.repository.CreateTopup(ctx, Topup{
		ID: id.String(), WalletID: wallet.ID, UserID: uid.String(), Reference: reference,
		AmountUSD: input.AmountUSD, Currency: "USD", QRPayload: qrPayload, QRMD5: qrMD5,
		Status: StatusPendingPayment, ExpiresAt: expiresAt,
	})
	if err != nil {
		return Topup{}, apperror.Wrap(err, "create wallet top-up")
	}
	return topup, nil
}

func (s *Service) GetTopup(ctx context.Context, userID, topupID string) (Topup, error) {
	uid, err := validator.UUID(userID)
	if err != nil {
		return Topup{}, apperror.ErrUnauthorized
	}
	id, err := validator.UUID(topupID)
	if err != nil {
		return Topup{}, apperror.NewValidation(map[string]string{"topup_id": "must be a valid UUID"})
	}
	topup, err := s.repository.GetTopup(ctx, uid, id)
	if err != nil {
		return Topup{}, err
	}
	return s.refreshTopup(ctx, topup)
}

type WebhookInput struct {
	Reference string `json:"reference"`
	QRMD5     string `json:"qr_md5"`
}

func (s *Service) HandleWebhook(ctx context.Context, input WebhookInput) (Topup, error) {
	var topup Topup
	var err error
	if strings.TrimSpace(input.Reference) != "" {
		topup, err = s.repository.GetTopupByReference(ctx, strings.TrimSpace(input.Reference))
	} else if strings.TrimSpace(input.QRMD5) != "" {
		topup, err = s.repository.GetTopupByMD5(ctx, strings.TrimSpace(input.QRMD5))
	} else {
		return Topup{}, apperror.NewValidation(map[string]string{"reference": "reference or qr_md5 is required"})
	}
	if err != nil {
		return Topup{}, err
	}
	return s.refreshTopup(ctx, topup)
}

func (s *Service) AdminTopups(ctx context.Context, search, status string, page, pageSize int) (AdminTopupResult, error) {
	page, pageSize = normalizePagination(page, pageSize)
	status = strings.TrimSpace(status)
	allowed := map[string]bool{"": true, StatusPendingPayment: true, StatusConfirmed: true, StatusFailed: true, StatusExpired: true, StatusReversed: true}
	if !allowed[status] {
		return AdminTopupResult{}, apperror.NewValidation(map[string]string{"status": "unsupported wallet top-up status"})
	}
	result, err := s.repository.ListTopupsAdmin(ctx, strings.TrimSpace(search), status, page, pageSize)
	if err != nil {
		return AdminTopupResult{}, apperror.Wrap(err, "load wallet top-ups")
	}
	return result, nil
}

func (s *Service) AdminRecheckTopup(ctx context.Context, topupID string) (Topup, error) {
	id, err := validator.UUID(topupID)
	if err != nil {
		return Topup{}, apperror.NewValidation(map[string]string{"topup_id": "must be a valid UUID"})
	}
	topup, err := s.repository.GetTopupByID(ctx, id)
	if err != nil {
		return Topup{}, err
	}
	return s.refreshTopup(ctx, topup)
}

func (s *Service) refreshTopup(ctx context.Context, topup Topup) (Topup, error) {
	if topup.Status != StatusPendingPayment {
		return topup, nil
	}
	now := s.now()
	if !topup.ExpiresAt.After(now) {
		if err := s.repository.ExpireTopup(ctx, uuid.MustParse(topup.ID)); err != nil {
			return Topup{}, apperror.Wrap(err, "expire wallet top-up")
		}
		topup.Status = StatusExpired
		topup.FailureReason = "QR expired before payment was verified"
		return topup, nil
	}
	if s.provider == nil {
		return topup, nil
	}
	payment, err := s.provider.CheckTransaction(ctx, topup.QRMD5)
	if errors.Is(err, ErrPaymentNotFound) {
		return topup, nil
	}
	if err != nil {
		return Topup{}, apperror.Wrap(err, "verify Bakong payment")
	}
	if err := s.validatePayment(topup, payment); err != nil {
		_ = s.repository.FailTopup(ctx, uuid.MustParse(topup.ID), err.Error())
		return Topup{}, err
	}
	confirmed, _, err := s.repository.ConfirmTopup(ctx, uuid.MustParse(topup.ID), payment.Hash, now)
	if err != nil {
		return Topup{}, apperror.Wrap(err, "credit wallet from Bakong payment")
	}
	return confirmed, nil
}

func (s *Service) validateAmount(amount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount < s.config.BakongMinTopupUSD || amount > s.config.BakongMaxTopupUSD {
		return apperror.NewValidation(map[string]string{"amount_usd": "must be within the configured top-up limits"})
	}
	if math.Abs(amount-math.Round(amount*100)/100) > 1e-9 {
		return apperror.NewValidation(map[string]string{"amount_usd": "must use no more than 2 decimal places"})
	}
	return nil
}

func (s *Service) validatePayment(topup Topup, payment Payment) error {
	if payment.Hash == "" || payment.ToAccount != s.config.BakongAccountID {
		return apperror.NewValidation(map[string]string{"payment": "Bakong payment recipient could not be verified"})
	}
	if strings.ToUpper(payment.Currency) != "USD" {
		return apperror.NewValidation(map[string]string{"payment": "Bakong payment currency does not match USD"})
	}
	if math.Abs(payment.Amount-topup.AmountUSD) > 0.005 {
		return apperror.NewValidation(map[string]string{"payment": "Bakong payment amount does not match the top-up"})
	}
	return nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
