package wallet

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
	EnsureWallet(ctx context.Context, userID uuid.UUID) (Wallet, error)
	GetWallet(ctx context.Context, userID uuid.UUID) (Wallet, error)
	ListLedger(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]LedgerEntry, bool, error)
	CreateTopup(ctx context.Context, topup Topup) (Topup, error)
	GetTopup(ctx context.Context, userID, topupID uuid.UUID) (Topup, error)
	GetTopupByID(ctx context.Context, topupID uuid.UUID) (Topup, error)
	ConfirmTopup(ctx context.Context, topupID uuid.UUID, providerTransactionID string, now time.Time) (Topup, bool, error)
	ExpireTopup(ctx context.Context, topupID uuid.UUID) error
	FailTopup(ctx context.Context, topupID uuid.UUID, reason string) error
	GetTopupByReference(ctx context.Context, reference string) (Topup, error)
	GetTopupByMD5(ctx context.Context, qrMD5 string) (Topup, error)
	ListTopupsAdmin(ctx context.Context, search, status string, page, pageSize int) (AdminTopupResult, error)
}

type sqlRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *sqlc.Queries, pool *pgxpool.Pool) Repository {
	return &sqlRepository{queries: queries, pool: pool}
}

func (r *sqlRepository) EnsureWallet(ctx context.Context, userID uuid.UUID) (Wallet, error) {
	row, err := r.queries.EnsureWallet(ctx, sqlc.EnsureWalletParams{ID: uuid.New(), UserID: userID})
	if err != nil {
		return Wallet{}, fmt.Errorf("ensure wallet: %w", err)
	}
	return walletFromEnsureRow(row), nil
}

func (r *sqlRepository) GetWallet(ctx context.Context, userID uuid.UUID) (Wallet, error) {
	row, err := r.queries.GetWallet(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Wallet{}, apperror.ErrNotFound
	}
	if err != nil {
		return Wallet{}, fmt.Errorf("get wallet: %w", err)
	}
	return walletFromGetRow(row), nil
}

func (r *sqlRepository) ListLedger(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]LedgerEntry, bool, error) {
	rows, err := r.queries.ListWalletLedger(ctx, sqlc.ListWalletLedgerParams{UserID: userID, Limit: int32(pageSize + 1), Offset: int32((page - 1) * pageSize)})
	if err != nil {
		return nil, false, fmt.Errorf("list wallet ledger: %w", err)
	}
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	entries := make([]LedgerEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, LedgerEntry{
			ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(),
			EntryType: row.EntryType, Reference: row.Reference,
			AvailableDeltaUSD: row.AvailableDeltaUsd, HeldDeltaUSD: row.HeldDeltaUsd,
			AvailableBalanceUSD: row.AvailableBalanceUsd, HeldBalanceUSD: row.HeldBalanceUsd,
			Description: row.Description, CreatedAt: row.CreatedAt.Time,
		})
	}
	return entries, hasMore, nil
}

func (r *sqlRepository) CreateTopup(ctx context.Context, topup Topup) (Topup, error) {
	id, err := uuid.Parse(topup.ID)
	if err != nil {
		return Topup{}, fmt.Errorf("parse top-up ID: %w", err)
	}
	walletID, err := uuid.Parse(topup.WalletID)
	if err != nil {
		return Topup{}, fmt.Errorf("parse wallet ID: %w", err)
	}
	userID, err := uuid.Parse(topup.UserID)
	if err != nil {
		return Topup{}, fmt.Errorf("parse user ID: %w", err)
	}
	row, err := r.queries.CreateWalletTopup(ctx, sqlc.CreateWalletTopupParams{
		ID: id, WalletID: walletID, UserID: userID, Reference: topup.Reference,
		Column5: topup.AmountUSD, Column6: topup.QRPayload, Column7: topup.QRMD5,
		Column8: timestamptz(topup.ExpiresAt),
	})
	if err != nil {
		return Topup{}, fmt.Errorf("create wallet top-up: %w", err)
	}
	return topupFromCreateRow(row), nil
}

func (r *sqlRepository) GetTopup(ctx context.Context, userID, topupID uuid.UUID) (Topup, error) {
	row, err := r.queries.GetWalletTopup(ctx, sqlc.GetWalletTopupParams{ID: topupID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Topup{}, apperror.ErrNotFound
	}
	if err != nil {
		return Topup{}, fmt.Errorf("get wallet top-up: %w", err)
	}
	return topupFromGetRow(row), nil
}

func (r *sqlRepository) GetTopupByID(ctx context.Context, topupID uuid.UUID) (Topup, error) {
	row, err := r.queries.GetWalletTopupByID(ctx, topupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Topup{}, apperror.ErrNotFound
	}
	if err != nil {
		return Topup{}, fmt.Errorf("get wallet top-up by ID: %w", err)
	}
	return topupFromIDRow(row), nil
}

func (r *sqlRepository) GetTopupByReference(ctx context.Context, reference string) (Topup, error) {
	row, err := r.queries.GetWalletTopupByReference(ctx, reference)
	if errors.Is(err, pgx.ErrNoRows) {
		return Topup{}, apperror.ErrNotFound
	}
	if err != nil {
		return Topup{}, fmt.Errorf("get wallet top-up by reference: %w", err)
	}
	return topupFromReferenceRow(row), nil
}

func (r *sqlRepository) GetTopupByMD5(ctx context.Context, qrMD5 string) (Topup, error) {
	row, err := r.queries.GetWalletTopupByMD5(ctx, qrMD5)
	if errors.Is(err, pgx.ErrNoRows) {
		return Topup{}, apperror.ErrNotFound
	}
	if err != nil {
		return Topup{}, fmt.Errorf("get wallet top-up by QR hash: %w", err)
	}
	return topupFromMD5Row(row), nil
}

func (r *sqlRepository) ConfirmTopup(ctx context.Context, topupID uuid.UUID, providerTransactionID string, now time.Time) (Topup, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Topup{}, false, fmt.Errorf("begin top-up confirmation: %w", err)
	}
	q := r.queries.WithTx(tx)
	row, err := q.GetWalletTopupForUpdate(ctx, topupID)
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return Topup{}, false, apperror.ErrNotFound
		}
		return Topup{}, false, fmt.Errorf("lock wallet top-up: %w", err)
	}
	topup := topupFromForUpdateRow(row)
	if topup.Status != StatusPendingPayment {
		if err := tx.Commit(ctx); err != nil {
			return Topup{}, false, fmt.Errorf("commit unchanged top-up: %w", err)
		}
		return topup, false, nil
	}
	if !topup.ExpiresAt.After(now) {
		if err := q.MarkWalletTopupExpired(ctx, topupID); err != nil {
			_ = tx.Rollback(ctx)
			return Topup{}, false, fmt.Errorf("expire wallet top-up: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return Topup{}, false, fmt.Errorf("commit expired top-up: %w", err)
		}
		topup.Status = StatusExpired
		topup.FailureReason = "QR expired before payment was verified"
		return topup, false, nil
	}
	walletRow, err := q.GetWalletForUpdate(ctx, row.WalletID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return Topup{}, false, fmt.Errorf("lock wallet for top-up: %w", err)
	}
	updatedWallet, err := q.CreditWallet(ctx, sqlc.CreditWalletParams{ID: row.WalletID, Column2: row.AmountUsd})
	if err != nil {
		_ = tx.Rollback(ctx)
		return Topup{}, false, fmt.Errorf("credit wallet: %w", err)
	}
	if err := q.InsertWalletLedgerEntry(ctx, sqlc.InsertWalletLedgerEntryParams{
		ID: uuid.New(), WalletID: row.WalletID, UserID: row.UserID, Column4: "topup_confirmed", Column5: row.Reference,
		Column6: row.AmountUsd, Column7: 0, Column8: updatedWallet.AvailableAmount, Column9: walletRow.HeldAmount,
		Column10: "Bakong wallet top-up confirmed",
	}); err != nil {
		_ = tx.Rollback(ctx)
		return Topup{}, false, fmt.Errorf("write wallet ledger entry: %w", err)
	}
	confirmed, err := q.ConfirmWalletTopup(ctx, sqlc.ConfirmWalletTopupParams{ID: topupID, Column2: providerTransactionID})
	if err != nil {
		_ = tx.Rollback(ctx)
		return Topup{}, false, fmt.Errorf("confirm wallet top-up: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Topup{}, false, fmt.Errorf("commit wallet top-up: %w", err)
	}
	return topupFromConfirmRow(confirmed), true, nil
}

func (r *sqlRepository) ExpireTopup(ctx context.Context, topupID uuid.UUID) error {
	if err := r.queries.MarkWalletTopupExpired(ctx, topupID); err != nil {
		return fmt.Errorf("expire wallet top-up: %w", err)
	}
	return nil
}

func (r *sqlRepository) FailTopup(ctx context.Context, topupID uuid.UUID, reason string) error {
	if err := r.queries.MarkWalletTopupFailed(ctx, sqlc.MarkWalletTopupFailedParams{ID: topupID, Column2: reason}); err != nil {
		return fmt.Errorf("fail wallet top-up: %w", err)
	}
	return nil
}

func (r *sqlRepository) ListTopupsAdmin(ctx context.Context, search, status string, page, pageSize int) (AdminTopupResult, error) {
	rows, err := r.queries.ListWalletTopupsAdmin(ctx, sqlc.ListWalletTopupsAdminParams{Column1: search, Column2: status, Limit: int32(pageSize), Offset: int32((page - 1) * pageSize)})
	if err != nil {
		return AdminTopupResult{}, fmt.Errorf("list wallet top-ups: %w", err)
	}
	total, err := r.queries.CountWalletTopupsAdmin(ctx, sqlc.CountWalletTopupsAdminParams{Column1: search, Column2: status})
	if err != nil {
		return AdminTopupResult{}, fmt.Errorf("count wallet top-ups: %w", err)
	}
	topups := make([]AdminTopup, 0, len(rows))
	for _, row := range rows {
		topup := Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference,
			AmountUSD: row.AmountUsd, Currency: row.Currency, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID),
			Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time,
			PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
		topups = append(topups, AdminTopup{Topup: topup, UserName: row.UserName, UserEmail: row.UserEmail})
	}
	return AdminTopupResult{Topups: topups, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages(total, pageSize)}, nil
}

func walletFromEnsureRow(row sqlc.EnsureWalletRow) Wallet {
	return Wallet{ID: row.ID.String(), UserID: row.UserID.String(), Currency: row.Currency, AvailableAmount: row.AvailableAmount, HeldAmount: row.HeldAmount, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func walletFromGetRow(row sqlc.GetWalletRow) Wallet {
	return Wallet{ID: row.ID.String(), UserID: row.UserID.String(), Currency: row.Currency, AvailableAmount: row.AvailableAmount, HeldAmount: row.HeldAmount, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func topupFromCreateRow(row sqlc.CreateWalletTopupRow) Topup {
	return Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference, AmountUSD: row.AmountUsd, Currency: row.Currency, QRPayload: row.QrPayload, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID), Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time, PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func topupFromGetRow(row sqlc.GetWalletTopupRow) Topup {
	return Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference, AmountUSD: row.AmountUsd, Currency: row.Currency, QRPayload: row.QrPayload, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID), Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time, PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func topupFromIDRow(row sqlc.GetWalletTopupByIDRow) Topup {
	return Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference, AmountUSD: row.AmountUsd, Currency: row.Currency, QRPayload: row.QrPayload, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID), Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time, PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func topupFromForUpdateRow(row sqlc.GetWalletTopupForUpdateRow) Topup {
	return Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference, AmountUSD: row.AmountUsd, Currency: row.Currency, QRPayload: row.QrPayload, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID), Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time, PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func topupFromReferenceRow(row sqlc.GetWalletTopupByReferenceRow) Topup {
	return Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference, AmountUSD: row.AmountUsd, Currency: row.Currency, QRPayload: row.QrPayload, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID), Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time, PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func topupFromMD5Row(row sqlc.GetWalletTopupByMD5Row) Topup {
	return Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference, AmountUSD: row.AmountUsd, Currency: row.Currency, QRPayload: row.QrPayload, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID), Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time, PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func topupFromConfirmRow(row sqlc.ConfirmWalletTopupRow) Topup {
	return Topup{ID: row.ID.String(), WalletID: row.WalletID.String(), UserID: row.UserID.String(), Reference: row.Reference, AmountUSD: row.AmountUsd, Currency: row.Currency, QRPayload: row.QrPayload, QRMD5: row.QrMd5, ProviderTransactionID: nullableText(row.ProviderTransactionID), Status: row.Status, FailureReason: row.FailureReason, ResolutionNote: row.ResolutionNote, ExpiresAt: row.ExpiresAt.Time, PaidAt: nullableTime(row.PaidAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func nullableText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func totalPages(total int64, pageSize int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}
