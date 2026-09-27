package facebook

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/validator"
)

type Service struct {
	repository      Repository
	pageRepo        PageRepository
	marketplaceRepo MarketplaceRepository
	client          *Client
	cipher          *TokenCipher
	stateStore      StateStore
	frontendURL     string
	logger          *slog.Logger
	now             func() time.Time
}

func NewService(repository Repository, client *Client, cipher *TokenCipher, stateStore StateStore, frontendURL string, logger *slog.Logger) *Service {
	pageRepo, _ := repository.(PageRepository)
	marketplaceRepo, _ := repository.(MarketplaceRepository)
	return &Service{
		repository: repository, pageRepo: pageRepo, marketplaceRepo: marketplaceRepo, client: client, cipher: cipher, stateStore: stateStore,
		frontendURL: frontendURL, logger: logger, now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Connect(ctx context.Context, userID, sessionID string) (string, error) {
	if !s.enabled() {
		return "", apperror.ErrDependencyUnavailable
	}
	uid, err := validator.UUID(userID)
	if err != nil || sessionID == "" {
		return "", apperror.ErrUnauthorized
	}
	state, err := newOAuthState()
	if err != nil {
		return "", apperror.Wrap(err, "generate Facebook OAuth state")
	}
	payload, err := encodeOAuthState(oauthState{UserID: uid.String(), SessionID: sessionID})
	if err != nil {
		return "", apperror.Wrap(err, "encode Facebook OAuth state")
	}
	if err := s.stateStore.Save(ctx, stateKey(state), payload, oauthStateTTL); err != nil {
		s.logger.ErrorContext(ctx, "Facebook connection state could not be stored", "error", err)
		return "", apperror.ErrDependencyUnavailable
	}
	authorizationURL, err := s.client.AuthorizationURL(state)
	if err != nil {
		return "", apperror.Wrap(err, "create Facebook authorization URL")
	}
	s.logger.InfoContext(ctx, "Facebook connection started", "user_id", uid.String())
	return authorizationURL, nil
}

func (s *Service) Callback(ctx context.Context, params CallbackParams) (CallbackStatus, error) {
	if !s.enabled() {
		return CallbackError, apperror.ErrDependencyUnavailable
	}
	if len(params.State) < 32 || len(params.State) > 256 {
		return CallbackStateErr, nil
	}
	statePayload, err := s.stateStore.Consume(ctx, stateKey(params.State))
	if err != nil {
		if errors.Is(err, errStateNotFound) {
			return CallbackStateErr, nil
		}
		s.logger.ErrorContext(ctx, "Facebook OAuth state could not be consumed", "error", err)
		return CallbackError, apperror.ErrDependencyUnavailable
	}
	state, err := decodeOAuthState(statePayload)
	if err != nil {
		return CallbackStateErr, nil
	}
	if params.OAuthError != "" {
		if params.OAuthError == "access_denied" || params.ErrorReason == "user_denied" {
			return CallbackCancelled, nil
		}
		return CallbackError, nil
	}
	if params.Code == "" || len(params.Code) > 4096 {
		return CallbackError, nil
	}
	userID, err := validator.UUID(state.UserID)
	if err != nil {
		return CallbackStateErr, nil
	}
	s.logger.InfoContext(ctx, "Facebook callback received", "user_id", userID.String())
	token, err := s.client.ExchangeAuthorizationCode(ctx, params.Code)
	if err != nil {
		s.logger.ErrorContext(ctx, "Facebook token exchange failed", "user_id", userID.String(), "error", err)
		return CallbackError, nil
	}
	profile, err := s.client.GetProfile(ctx, token.AccessToken)
	if err != nil {
		s.logger.ErrorContext(ctx, "Facebook profile lookup failed", "user_id", userID.String(), "error", err)
		return CallbackError, nil
	}
	encryptedToken, err := s.cipher.encrypt(token.AccessToken)
	if err != nil {
		return CallbackError, apperror.Wrap(err, "encrypt Facebook access token")
	}
	var expiresAt *time.Time
	if token.ExpiresIn > 0 {
		expires := s.now().Add(time.Duration(token.ExpiresIn) * time.Second)
		expiresAt = &expires
	}
	connection, err := s.repository.Upsert(ctx, userID, profile.ID, profile.Name, encryptedToken, expiresAt, requestedScope)
	if err != nil {
		s.logger.ErrorContext(ctx, "Facebook connection could not be saved", "user_id", userID.String(), "error", err)
		return CallbackError, apperror.Wrap(err, "save Facebook connection")
	}
	if _, err := s.syncConnection(ctx, userID.String(), connection.ID); err != nil {
		s.logger.WarnContext(ctx, "Facebook Pages could not be synced after connection", "user_id", userID.String(), "error", err)
	}
	s.logger.InfoContext(ctx, "Facebook account connected", "user_id", userID.String(), "facebook_account_id", profile.ID)
	return CallbackConnected, nil
}

func (s *Service) Connections(ctx context.Context, userID string) ([]Connection, error) {
	uid, err := validator.UUID(userID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}
	connections, err := s.repository.ListActive(ctx, uid)
	if err != nil {
		return nil, apperror.Wrap(err, "list Facebook connections")
	}
	result := make([]Connection, 0, len(connections))
	for _, connection := range connections {
		result = append(result, connection.Connection)
	}
	return result, nil
}

func (s *Service) Pages(ctx context.Context, userID, connectionID string) ([]Page, error) {
	return s.syncConnection(ctx, userID, connectionID)
}

func (s *Service) syncConnection(ctx context.Context, userID, connectionID string) ([]Page, error) {
	if !s.enabled() {
		return nil, apperror.ErrDependencyUnavailable
	}
	uid, err := validator.UUID(userID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}
	cid, err := validator.UUID(connectionID)
	if err != nil {
		return nil, apperror.NewValidation(map[string]string{"connection_id": "must be a valid UUID"})
	}
	connection, err := s.repository.GetActive(ctx, uid, cid)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.Wrap(err, "load Facebook connection")
	}
	if connection.TokenExpiresAt != nil && !connection.TokenExpiresAt.After(s.now()) {
		_ = s.repository.Delete(ctx, uid, cid)
		return nil, &apperror.Error{Status: 409, Code: "FACEBOOK_REAUTH_REQUIRED", Message: "This Facebook connection needs to be reconnected"}
	}
	accessToken, err := s.cipher.decrypt(connection.EncryptedAccessToken)
	if err != nil {
		return nil, apperror.Wrap(err, "decrypt Facebook access token")
	}
	pages, err := s.client.ListPages(ctx, accessToken)
	if err != nil {
		if IsInvalidToken(err) || IsPermissionError(err) {
			_ = s.repository.Delete(ctx, uid, cid)
			s.logger.WarnContext(ctx, "Facebook connection requires reauthorization", "user_id", uid.String(), "facebook_account_id", connection.FacebookUserID)
			return nil, &apperror.Error{Status: 409, Code: "FACEBOOK_REAUTH_REQUIRED", Message: "This Facebook connection needs to be reconnected"}
		}
		return nil, apperror.Wrap(err, "retrieve Facebook Pages")
	}
	if s.pageRepo != nil {
		for _, page := range pages {
			pageID, err := s.pageRepo.UpsertPage(ctx, page.ID, page.Name)
			if err != nil {
				return nil, apperror.Wrap(err, "store Facebook Page")
			}
			if err := s.pageRepo.LinkPage(ctx, pageID, cid); err != nil {
				return nil, apperror.Wrap(err, "link Facebook Page")
			}
		}
	}
	s.logger.InfoContext(ctx, "Facebook Pages retrieved", "user_id", uid.String(), "facebook_account_id", connection.FacebookUserID, "page_count", len(pages))
	return pages, nil
}

func (s *Service) SyncAll(ctx context.Context, userID string) (int, error) {
	connections, err := s.Connections(ctx, userID)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, connection := range connections {
		pages, syncErr := s.syncConnection(ctx, userID, connection.ID)
		if syncErr != nil {
			return total, syncErr
		}
		total += len(pages)
	}
	return total, nil
}

func (s *Service) SyncConnection(ctx context.Context, userID, connectionID string) ([]Page, error) {
	return s.syncConnection(ctx, userID, connectionID)
}

func (s *Service) StoredPages(ctx context.Context, userID string, params pageInventoryParams) (PageInventory, error) {
	if s.pageRepo == nil {
		return PageInventory{}, apperror.ErrDependencyUnavailable
	}
	uid, err := validator.UUID(userID)
	if err != nil {
		return PageInventory{}, apperror.ErrUnauthorized
	}
	params.UserID = uid
	params.Page, params.PageSize = normalizePagination(params.Page, params.PageSize)
	pages, total, err := s.pageRepo.ListPageInventory(ctx, params)
	if err != nil {
		return PageInventory{}, apperror.Wrap(err, "list stored Facebook Pages")
	}
	return PageInventory{Pages: pages, Total: total, Page: params.Page, PageSize: params.PageSize, TotalPages: totalPages(total, params.PageSize)}, nil
}

func (s *Service) SaveListing(ctx context.Context, userID, pageID string, input ListingInput) (Listing, error) {
	if s.pageRepo == nil {
		return Listing{}, apperror.ErrDependencyUnavailable
	}
	uid, err := validator.UUID(userID)
	if err != nil {
		return Listing{}, apperror.ErrUnauthorized
	}
	pid, err := validator.UUID(pageID)
	if err != nil {
		return Listing{}, apperror.NewValidation(map[string]string{"page_id": "must be a valid UUID"})
	}
	if err := validateListingInput(&input); err != nil {
		return Listing{}, err
	}
	if err := s.pageRepo.GetPage(ctx, uid, pid); err != nil {
		return Listing{}, err
	}
	listing, err := s.pageRepo.UpsertListing(ctx, uid, pid, input)
	if err != nil {
		return Listing{}, apperror.Wrap(err, "save Facebook Page listing")
	}
	return listing, nil
}

func (s *Service) CreateOrder(ctx context.Context, userID, listingID string, input MarketplaceOrderInput) (MarketplaceOrder, error) {
	if s.marketplaceRepo == nil {
		return MarketplaceOrder{}, apperror.ErrDependencyUnavailable
	}
	uid, err := validator.UUID(userID)
	if err != nil {
		return MarketplaceOrder{}, apperror.ErrUnauthorized
	}
	id, err := validator.UUID(listingID)
	if err != nil {
		return MarketplaceOrder{}, apperror.NewValidation(map[string]string{"listing_id": "must be a valid UUID"})
	}
	if err := validateMarketplaceOrderInput(&input); err != nil {
		return MarketplaceOrder{}, err
	}
	if err := s.marketplaceRepo.ReleaseExpiredReservations(ctx); err != nil {
		return MarketplaceOrder{}, apperror.Wrap(err, "release expired marketplace reservations")
	}
	referenceID, err := uuid.NewV7()
	if err != nil {
		return MarketplaceOrder{}, apperror.Wrap(err, "generate marketplace order reference")
	}
	reference := "OS-" + strings.ToUpper(referenceID.String()[:8])
	expiresAt := s.now().Add(reservationTTL)
	var order MarketplaceOrder
	if input.PaymentMethod == "wallet" {
		order, err = s.marketplaceRepo.CreateWalletMarketplaceOrder(ctx, id, reference, uid, input, expiresAt)
	} else {
		order, err = s.marketplaceRepo.CreateMarketplaceOrder(ctx, id, reference, uid, input, expiresAt)
	}
	if errors.Is(err, apperror.ErrInsufficientFunds) {
		return MarketplaceOrder{}, err
	}
	if errors.Is(err, apperror.ErrConflict) {
		return MarketplaceOrder{}, &apperror.Error{Status: 409, Code: "LISTING_UNAVAILABLE", Message: "This listing is no longer available"}
	}
	if err != nil {
		return MarketplaceOrder{}, apperror.Wrap(err, "create marketplace order")
	}
	return order, nil
}

func (s *Service) PublishedListings(ctx context.Context, search string, page, pageSize int) (PageInventory, error) {
	if s.pageRepo == nil {
		return PageInventory{}, apperror.ErrDependencyUnavailable
	}
	if s.marketplaceRepo != nil {
		if err := s.marketplaceRepo.ReleaseExpiredReservations(ctx); err != nil {
			return PageInventory{}, apperror.Wrap(err, "release expired marketplace reservations")
		}
	}
	page, pageSize = normalizePagination(page, pageSize)
	listings, total, err := s.pageRepo.ListPublishedListings(ctx, trimOAuthValue(search), page, pageSize)
	if err != nil {
		return PageInventory{}, apperror.Wrap(err, "list published Facebook Page listings")
	}
	pages := make([]PageInventoryItem, 0, len(listings))
	for _, listing := range listings {
		pages = append(pages, PageInventoryItem{
			ID: listing.PageID, FacebookPageID: listing.FacebookPageID, Name: listing.PageName,
			Listing: &listing.Listing,
		})
	}
	return PageInventory{Pages: pages, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages(total, pageSize)}, nil
}

func (s *Service) PublishedListing(ctx context.Context, listingID string) (PublicListing, error) {
	if s.pageRepo == nil {
		return PublicListing{}, apperror.ErrDependencyUnavailable
	}
	if s.marketplaceRepo != nil {
		if err := s.marketplaceRepo.ReleaseExpiredReservations(ctx); err != nil {
			return PublicListing{}, apperror.Wrap(err, "release expired marketplace reservations")
		}
	}
	id, err := validator.UUID(listingID)
	if err != nil {
		return PublicListing{}, apperror.NewValidation(map[string]string{"listing_id": "must be a valid UUID"})
	}
	listing, err := s.pageRepo.GetPublishedListing(ctx, id)
	if err != nil {
		return PublicListing{}, err
	}
	return listing, nil
}

func (s *Service) Disconnect(ctx context.Context, userID, connectionID string) error {
	uid, err := validator.UUID(userID)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	cid, err := validator.UUID(connectionID)
	if err != nil {
		return apperror.NewValidation(map[string]string{"connection_id": "must be a valid UUID"})
	}
	if err := s.repository.Delete(ctx, uid, cid); err != nil {
		return apperror.Wrap(err, "disconnect Facebook account")
	}
	s.logger.InfoContext(ctx, "Facebook disconnected", "user_id", uid.String(), "connection_id", cid.String())
	return nil
}

func (s *Service) redirectURL(status CallbackStatus) string {
	frontend, err := url.Parse(s.frontendURL)
	if err != nil || frontend.Scheme == "" || frontend.Host == "" {
		return "http://localhost:3000/?view=facebook&status=error"
	}
	query := frontend.Query()
	query.Set("view", "facebook")
	query.Set("status", string(status))
	frontend.RawQuery = query.Encode()
	return frontend.String()
}

func (s *Service) enabled() bool {
	return s.client != nil && s.cipher != nil && s.stateStore != nil && s.repository != nil
}

func validateListingInput(input *ListingInput) error {
	fields := make(map[string]string)
	input.Title = strings.TrimSpace(input.Title)
	input.Subtitle = strings.TrimSpace(input.Subtitle)
	input.Description = strings.TrimSpace(input.Description)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.DeliveryWindow = strings.TrimSpace(input.DeliveryWindow)
	if input.Title == "" || len(input.Title) > 255 {
		fields["title"] = "must be between 1 and 255 characters"
	}
	if input.Subtitle == "" || len(input.Subtitle) > 255 {
		fields["subtitle"] = "must be between 1 and 255 characters"
	}
	if len(input.Description) > 5000 {
		fields["description"] = "must be at most 5000 characters"
	}
	if math.IsNaN(input.PriceUSD) || math.IsInf(input.PriceUSD, 0) || input.PriceUSD < 0 {
		fields["price_usd"] = "must be a valid USD amount of zero or greater"
	} else if math.Abs(input.PriceUSD-math.Round(input.PriceUSD*100)/100) > 1e-9 {
		fields["price_usd"] = "must use no more than 2 decimal places"
	} else {
		input.PriceUSD = math.Round(input.PriceUSD*100) / 100
	}
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	if input.Currency != "USD" {
		fields["currency"] = "must be USD"
	}
	if input.DeliveryWindow == "" || len(input.DeliveryWindow) > 120 {
		fields["delivery_window"] = "must be between 1 and 120 characters"
	}
	switch input.Status {
	case "draft", "pending_review", "published", "unlisted":
	default:
		fields["status"] = "must be draft, pending_review, published, or unlisted"
	}
	if input.SortOrder < 0 {
		fields["sort_order"] = "must be zero or greater"
	}
	if len(fields) > 0 {
		return apperror.NewValidation(fields)
	}
	return nil
}

func validateMarketplaceOrderInput(input *MarketplaceOrderInput) error {
	fields := make(map[string]string)
	input.BuyerNote = strings.TrimSpace(input.BuyerNote)
	input.PaymentMethod = strings.ToLower(strings.TrimSpace(input.PaymentMethod))
	if input.PaymentMethod == "" {
		input.PaymentMethod = "wallet"
	}
	if input.PaymentMethod != "wallet" {
		fields["payment_method"] = "wallet is currently the only supported Page order payment method"
	}
	if len(input.BuyerNote) > 2000 {
		fields["buyer_note"] = "must be at most 2000 characters"
	}
	if !input.TermsAccepted {
		fields["terms_accepted"] = "must be accepted before reserving a listing"
	}
	if len(fields) > 0 {
		return apperror.NewValidation(fields)
	}
	return nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func totalPages(total int64, pageSize int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

func trimOAuthValue(value string) string { return strings.TrimSpace(value) }
