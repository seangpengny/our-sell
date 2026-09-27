package facebook

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/vtech/our-sell/db/sqlc"
	"github.com/vtech/our-sell/internal/platform/apperror"
)

type Repository interface {
	Upsert(ctx context.Context, userID uuid.UUID, facebookUserID, name string, encryptedToken []byte, expiresAt *time.Time, scopes string) (storedConnection, error)
	ListActive(ctx context.Context, userID uuid.UUID) ([]storedConnection, error)
	GetActive(ctx context.Context, userID, connectionID uuid.UUID) (storedConnection, error)
	Delete(ctx context.Context, userID, connectionID uuid.UUID) error
}

type PageRepository interface {
	UpsertPage(ctx context.Context, pageID, name string) (uuid.UUID, error)
	LinkPage(ctx context.Context, pageID, connectionID uuid.UUID) error
	ListPageInventory(ctx context.Context, params pageInventoryParams) ([]PageInventoryItem, int64, error)
	GetPage(ctx context.Context, userID, pageID uuid.UUID) error
	UpsertListing(ctx context.Context, userID, pageID uuid.UUID, input ListingInput) (Listing, error)
	ListPublishedListings(ctx context.Context, search string, page, pageSize int) ([]PublicListing, int64, error)
	GetPublishedListing(ctx context.Context, listingID uuid.UUID) (PublicListing, error)
}

type MarketplaceRepository interface {
	ReleaseExpiredReservations(ctx context.Context) error
	CreateMarketplaceOrder(ctx context.Context, listingID uuid.UUID, reference string, buyerUserID uuid.UUID, input MarketplaceOrderInput, expiresAt time.Time) (MarketplaceOrder, error)
	CreateWalletMarketplaceOrder(ctx context.Context, listingID uuid.UUID, reference string, buyerUserID uuid.UUID, input MarketplaceOrderInput, expiresAt time.Time) (MarketplaceOrder, error)
}

type sqlRepository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) Upsert(ctx context.Context, userID uuid.UUID, facebookUserID, name string, encryptedToken []byte, expiresAt *time.Time, scopes string) (storedConnection, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return storedConnection{}, fmt.Errorf("generate Facebook connection ID: %w", err)
	}
	row, err := r.queries.UpsertFacebookConnection(ctx, sqlc.UpsertFacebookConnectionParams{
		ID:                   id,
		UserID:               userID,
		FacebookUserID:       facebookUserID,
		FacebookUserName:     name,
		EncryptedAccessToken: encryptedToken,
		TokenExpiresAt:       nullableTimestamp(expiresAt),
		Scopes:               scopes,
	})
	if err != nil {
		return storedConnection{}, fmt.Errorf("upsert Facebook connection: %w", err)
	}
	return fromSQLConnection(row), nil
}

func (r *sqlRepository) ListActive(ctx context.Context, userID uuid.UUID) ([]storedConnection, error) {
	rows, err := r.queries.ListFacebookConnections(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list Facebook connections: %w", err)
	}
	result := make([]storedConnection, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromSQLConnection(row))
	}
	return result, nil
}

func (r *sqlRepository) GetActive(ctx context.Context, userID, connectionID uuid.UUID) (storedConnection, error) {
	row, err := r.queries.GetFacebookConnection(ctx, sqlc.GetFacebookConnectionParams{ID: connectionID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return storedConnection{}, apperror.ErrNotFound
	}
	if err != nil {
		return storedConnection{}, fmt.Errorf("get Facebook connection: %w", err)
	}
	return fromSQLConnection(row), nil
}

func (r *sqlRepository) Delete(ctx context.Context, userID, connectionID uuid.UUID) error {
	if err := r.queries.DeleteFacebookConnection(ctx, sqlc.DeleteFacebookConnectionParams{ID: connectionID, UserID: userID}); err != nil {
		return fmt.Errorf("delete Facebook connection: %w", err)
	}
	return nil
}

func (r *sqlRepository) UpsertPage(ctx context.Context, pageID, name string) (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate Facebook Page ID: %w", err)
	}
	row, err := r.queries.UpsertFacebookPage(ctx, sqlc.UpsertFacebookPageParams{
		ID: id, FacebookPageID: pageID, Name: name,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert Facebook Page: %w", err)
	}
	return row.ID, nil
}

func (r *sqlRepository) LinkPage(ctx context.Context, pageID, connectionID uuid.UUID) error {
	if err := r.queries.LinkFacebookPageToConnection(ctx, sqlc.LinkFacebookPageToConnectionParams{PageID: pageID, ConnectionID: connectionID}); err != nil {
		return fmt.Errorf("link Facebook Page to connection: %w", err)
	}
	return nil
}

func (r *sqlRepository) ListPageInventory(ctx context.Context, params pageInventoryParams) ([]PageInventoryItem, int64, error) {
	connectionID := params.ConnectionID
	rows, err := r.queries.ListFacebookPageInventory(ctx, sqlc.ListFacebookPageInventoryParams{
		UserID: params.UserID, Column2: connectionID, Column3: params.Search,
		Limit: int32(params.PageSize), Offset: int32((params.Page - 1) * params.PageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list Facebook Page inventory: %w", err)
	}
	total, err := r.queries.CountFacebookPageInventory(ctx, sqlc.CountFacebookPageInventoryParams{
		UserID: params.UserID, Column2: connectionID, Column3: params.Search,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count Facebook Page inventory: %w", err)
	}
	items := make([]PageInventoryItem, 0, len(rows))
	for _, row := range rows {
		item := PageInventoryItem{
			ID: row.ID.String(), FacebookPageID: row.FacebookPageID, Name: row.Name,
			ConnectionID: row.ConnectionID.String(), ConnectionName: row.ConnectionName,
			LastSyncedAt: nullableTime(row.LastSyncedAt),
		}
		if row.ListingID.Valid {
			item.Listing = &Listing{
				ID: uuid.UUID(row.ListingID.Bytes).String(), PageID: row.ID.String(),
				Title: nullableString(row.ListingTitle), Subtitle: nullableString(row.ListingSubtitle),
				Description: nullableString(row.ListingDescription), PriceUSD: row.PriceUsd,
				Currency: nullableString(row.Currency), DeliveryWindow: nullableString(row.DeliveryWindow),
				Status: nullableString(row.ListingStatus), Featured: nullableBool(row.Featured),
				SortOrder: nullableInt32(row.SortOrder), PublishedAt: nullableTime(row.PublishedAt),
			}
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *sqlRepository) GetPage(ctx context.Context, userID, pageID uuid.UUID) error {
	_, err := r.queries.GetFacebookPageForUser(ctx, sqlc.GetFacebookPageForUserParams{ID: pageID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get Facebook Page: %w", err)
	}
	return nil
}

func (r *sqlRepository) UpsertListing(ctx context.Context, userID, pageID uuid.UUID, input ListingInput) (Listing, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Listing{}, fmt.Errorf("generate Page listing ID: %w", err)
	}
	row, err := r.queries.UpsertFacebookPageListing(ctx, sqlc.UpsertFacebookPageListingParams{
		ID: id, PageID: pageID, SellerUserID: userID, Title: input.Title, Subtitle: input.Subtitle,
		Description: input.Description, Column7: input.PriceUSD, Currency: input.Currency,
		DeliveryWindow: input.DeliveryWindow, Status: input.Status, Featured: input.Featured,
		SortOrder: int32(input.SortOrder),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Listing{}, apperror.ErrConflict
	}
	if err != nil {
		return Listing{}, fmt.Errorf("upsert Page listing: %w", err)
	}
	return fromSQLListing(row), nil
}

func (r *sqlRepository) ReleaseExpiredReservations(ctx context.Context) error {
	if err := r.queries.ReleaseExpiredMarketplaceReservations(ctx); err != nil {
		return fmt.Errorf("release expired marketplace reservations: %w", err)
	}
	return nil
}

func (r *sqlRepository) CreateMarketplaceOrder(ctx context.Context, listingID uuid.UUID, reference string, buyerUserID uuid.UUID, input MarketplaceOrderInput, expiresAt time.Time) (MarketplaceOrder, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return MarketplaceOrder{}, fmt.Errorf("generate marketplace order ID: %w", err)
	}
	row, err := r.queries.CreateMarketplaceOrder(ctx, sqlc.CreateMarketplaceOrderParams{
		ID: id, ID_2: listingID, Reference: reference,
		BuyerUserID: pgtype.UUID{Bytes: buyerUserID, Valid: true},
		BuyerNote:   input.BuyerNote,
		Column6:     input.PaymentMethod, ReservationExpiresAt: pgtype.Timestamptz{Time: expiresAt.UTC(), Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketplaceOrder{}, apperror.ErrConflict
	}
	if err != nil {
		return MarketplaceOrder{}, fmt.Errorf("create marketplace order: %w", err)
	}
	return fromSQLMarketplaceOrder(row), nil
}

func (r *sqlRepository) CreateWalletMarketplaceOrder(ctx context.Context, listingID uuid.UUID, reference string, buyerUserID uuid.UUID, input MarketplaceOrderInput, expiresAt time.Time) (MarketplaceOrder, error) {
	price, err := r.queries.GetPublishedListingPrice(ctx, listingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketplaceOrder{}, apperror.ErrConflict
	}
	if err != nil {
		return MarketplaceOrder{}, fmt.Errorf("get wallet order price: %w", err)
	}
	available, err := r.queries.GetWalletAvailableBalance(ctx, buyerUserID)
	if errors.Is(err, pgx.ErrNoRows) || available+0.0000001 < price {
		return MarketplaceOrder{}, apperror.ErrInsufficientFunds
	}
	row, err := r.queries.CreateWalletMarketplaceOrder(ctx, sqlc.CreateWalletMarketplaceOrderParams{
		ID: uuid.New(), ID_2: listingID, Reference: reference, UserID: buyerUserID,
		BuyerNote: input.BuyerNote, ReservationExpiresAt: pgtype.Timestamptz{Time: expiresAt.UTC(), Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketplaceOrder{}, apperror.ErrInsufficientFunds
	}
	if err != nil {
		return MarketplaceOrder{}, fmt.Errorf("create wallet marketplace order: %w", err)
	}
	return MarketplaceOrder{ID: row.ID.String(), Reference: row.Reference, ListingID: row.ListingID.String(), BuyerUserID: uuid.UUID(row.BuyerUserID.Bytes).String(), BuyerNote: row.BuyerNote, AmountUSD: row.AmountUsd, Currency: row.Currency, PaymentMethod: row.PaymentMethod, Status: row.Status, ReservationExpiresAt: row.ReservationExpiresAt.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time, PaidAt: nullableTime(row.PaidAt), CompletedAt: nullableTime(row.CompletedAt)}, nil
}

func (r *sqlRepository) ListPublishedListings(ctx context.Context, search string, page, pageSize int) ([]PublicListing, int64, error) {
	rows, err := r.queries.ListPublishedFacebookPageListings(ctx, sqlc.ListPublishedFacebookPageListingsParams{
		Column1: search, Limit: int32(pageSize), Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list published Page listings: %w", err)
	}
	total, err := r.queries.CountPublishedFacebookPageListings(ctx, search)
	if err != nil {
		return nil, 0, fmt.Errorf("count published Page listings: %w", err)
	}
	items := make([]PublicListing, 0, len(rows))
	for _, row := range rows {
		items = append(items, PublicListing{
			Listing: Listing{ID: row.ID.String(), PageID: row.PageID.String(), Title: row.Title, Subtitle: row.Subtitle,
				Description: row.Description, PriceUSD: row.PriceUsd, Currency: row.Currency,
				DeliveryWindow: row.DeliveryWindow, Status: row.Status, Featured: row.Featured,
				SortOrder: int(row.SortOrder), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
				PublishedAt: nullableTime(row.PublishedAt)},
			FacebookPageID: row.FacebookPageID, PageName: row.Name,
		})
	}
	return items, total, nil
}

func (r *sqlRepository) GetPublishedListing(ctx context.Context, listingID uuid.UUID) (PublicListing, error) {
	row, err := r.queries.GetPublishedFacebookPageListing(ctx, listingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicListing{}, apperror.ErrNotFound
	}
	if err != nil {
		return PublicListing{}, fmt.Errorf("get published Page listing: %w", err)
	}
	return PublicListing{
		Listing: Listing{ID: row.ID.String(), PageID: row.PageID.String(), Title: row.Title, Subtitle: row.Subtitle,
			Description: row.Description, PriceUSD: row.PriceUsd, Currency: row.Currency,
			DeliveryWindow: row.DeliveryWindow, Status: row.Status, Featured: row.Featured,
			SortOrder: int(row.SortOrder), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
			PublishedAt: nullableTime(row.PublishedAt)},
		FacebookPageID: row.FacebookPageID, PageName: row.Name,
	}, nil
}

func fromSQLConnection(row sqlc.FacebookConnection) storedConnection {
	return storedConnection{
		Connection: Connection{
			ID:             row.ID.String(),
			FacebookUserID: row.FacebookUserID,
			Name:           row.FacebookUserName,
			ConnectedAt:    row.CreatedAt.Time,
			TokenExpiresAt: nullableTime(row.TokenExpiresAt),
		},
		UserID:               row.UserID.String(),
		EncryptedAccessToken: row.EncryptedAccessToken,
		Scopes:               row.Scopes,
	}
}

func nullableTimestamp(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func nullableString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableInt64(value pgtype.Int8) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func nullableInt32(value pgtype.Int4) int {
	if !value.Valid {
		return 0
	}
	return int(value.Int32)
}

func nullableBool(value pgtype.Bool) bool {
	return value.Valid && value.Bool
}

func fromSQLListing(row sqlc.UpsertFacebookPageListingRow) Listing {
	return Listing{ID: row.ID.String(), PageID: row.PageID.String(), Title: row.Title, Subtitle: row.Subtitle,
		Description: row.Description, PriceUSD: row.PriceUsd, Currency: row.Currency,
		DeliveryWindow: row.DeliveryWindow, Status: row.Status, Featured: row.Featured,
		SortOrder: int(row.SortOrder), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
		PublishedAt: nullableTime(row.PublishedAt)}
}

func fromSQLMarketplaceOrder(row sqlc.CreateMarketplaceOrderRow) MarketplaceOrder {
	buyerUserID := ""
	if row.BuyerUserID.Valid {
		buyerUserID = uuid.UUID(row.BuyerUserID.Bytes).String()
	}
	return MarketplaceOrder{
		ID: row.ID.String(), Reference: row.Reference, ListingID: row.ListingID.String(),
		BuyerUserID: buyerUserID, BuyerNote: row.BuyerNote,
		AmountUSD: row.AmountUsd, Currency: row.Currency, PaymentMethod: row.PaymentMethod,
		Status: row.Status, ReservationExpiresAt: row.ReservationExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
		PaidAt: nullableTime(row.PaidAt), CompletedAt: nullableTime(row.CompletedAt),
	}
}
