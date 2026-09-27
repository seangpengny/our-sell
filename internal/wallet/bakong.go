package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	khqr "github.com/ishinvin/go-khqr"

	"github.com/vtech/our-sell/internal/config"
)

var ErrPaymentNotFound = errors.New("Bakong payment was not found")

type Payment struct {
	Hash        string
	FromAccount string
	ToAccount   string
	Currency    string
	Amount      float64
	Description string
}

type Provider interface {
	GenerateQR(reference string, amountUSD float64, expiresAt time.Time) (qrPayload, qrMD5 string, err error)
	CheckTransaction(ctx context.Context, qrMD5 string) (Payment, error)
}

type bakongClient struct {
	apiToken   string
	baseURL    string
	accountID  string
	merchant   string
	city       string
	merchantID string
	bank       string
	storeLabel string
	httpClient *http.Client
}

func NewProvider(cfg config.Config) Provider {
	if cfg.BakongAPIToken == "" || cfg.BakongAccountID == "" {
		return nil
	}
	return &bakongClient{
		apiToken: cfg.BakongAPIToken, baseURL: cfg.BakongAPIBaseURL, accountID: cfg.BakongAccountID,
		merchant: cfg.BakongMerchantName, city: cfg.BakongMerchantCity, merchantID: cfg.BakongMerchantID,
		bank: cfg.BakongAcquiringBank, storeLabel: cfg.BakongStoreLabel,
		httpClient: &http.Client{Timeout: cfg.BakongHTTPTimeout},
	}
}

func (c *bakongClient) GenerateQR(reference string, amountUSD float64, expiresAt time.Time) (string, string, error) {
	if c.merchantID != "" && c.bank != "" {
		data, err := khqr.GenerateMerchant(khqr.MerchantInfo{
			BakongAccountID: c.accountID, MerchantName: c.merchant, MerchantCity: c.city,
			MerchantID: c.merchantID, AcquiringBank: c.bank, Currency: khqr.USD,
			Amount: amountUSD, ExpirationTimestamp: expiresAt.UnixMilli(), BillNumber: reference,
			StoreLabel: c.storeLabel,
		})
		if err != nil {
			return "", "", fmt.Errorf("generate merchant KHQR: %w", err)
		}
		return data.QR, data.MD5(), nil
	}
	data, err := khqr.GenerateIndividual(khqr.IndividualInfo{
		BakongAccountID: c.accountID, MerchantName: c.merchant, MerchantCity: c.city,
		Currency: khqr.USD, Amount: amountUSD, ExpirationTimestamp: expiresAt.UnixMilli(),
		BillNumber: reference, StoreLabel: c.storeLabel,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate KHQR: %w", err)
	}
	return data.QR, data.MD5(), nil
}

type transactionResponse struct {
	ResponseCode    int    `json:"responseCode"`
	ErrorCode       int    `json:"errorCode"`
	ResponseMessage string `json:"responseMessage"`
	Data            *struct {
		Hash          string  `json:"hash"`
		FromAccountID string  `json:"fromAccountId"`
		ToAccountID   string  `json:"toAccountId"`
		Currency      string  `json:"currency"`
		Amount        float64 `json:"amount"`
		Description   string  `json:"description"`
	} `json:"data"`
}

func (c *bakongClient) CheckTransaction(ctx context.Context, qrMD5 string) (Payment, error) {
	body, err := json.Marshal(struct {
		MD5 string `json:"md5"`
	}{MD5: qrMD5})
	if err != nil {
		return Payment{}, fmt.Errorf("encode Bakong status request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/check_transaction_by_md5", bytes.NewReader(body))
	if err != nil {
		return Payment{}, fmt.Errorf("create Bakong status request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.apiToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return Payment{}, fmt.Errorf("request Bakong status: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return Payment{}, ErrPaymentNotFound
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Payment{}, fmt.Errorf("Bakong status returned HTTP %d", response.StatusCode)
	}
	var payload transactionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return Payment{}, fmt.Errorf("decode Bakong status response: %w", err)
	}
	if payload.ResponseCode != 0 || payload.Data == nil {
		if payload.ResponseCode == 1 || payload.ErrorCode == 1 || strings.Contains(strings.ToLower(payload.ResponseMessage), "not found") {
			return Payment{}, ErrPaymentNotFound
		}
		return Payment{}, fmt.Errorf("Bakong status failed: %s", payload.ResponseMessage)
	}
	return Payment{
		Hash: payload.Data.Hash, FromAccount: payload.Data.FromAccountID, ToAccount: payload.Data.ToAccountID,
		Currency: payload.Data.Currency, Amount: payload.Data.Amount, Description: payload.Data.Description,
	}, nil
}
