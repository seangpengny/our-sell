package wallet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	khqr "github.com/ishinvin/go-khqr"

	"github.com/vtech/our-sell/internal/config"
)

func TestBakongClientGeneratesUSDQR(t *testing.T) {
	client := &bakongClient{
		accountID:  "oursell@bkrt",
		merchant:   "Our Sell",
		city:       "Phnom Penh",
		storeLabel: "Wallet",
	}

	qr, hash, err := client.GenerateQR("WT-ABC123", 10.25, time.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("GenerateQR() error = %v", err)
	}
	if qr == "" || hash == "" {
		t.Fatal("GenerateQR() returned an empty QR or MD5")
	}
	if err := khqr.Verify(qr); err != nil {
		t.Fatalf("generated QR failed verification: %v", err)
	}
	decoded, err := khqr.Decode(qr)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if decoded.BakongAccountID != client.accountID || decoded.TransactionCurrency != "840" || decoded.BillNumber != "WT-ABC123" {
		t.Fatalf("generated QR fields = account %q currency %q bill %q", decoded.BakongAccountID, decoded.TransactionCurrency, decoded.BillNumber)
	}
}

func TestBakongClientChecksTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/check_transaction_by_md5" {
			t.Fatalf("unexpected Bakong request: %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("authorization header = %q", request.Header.Get("Authorization"))
		}
		_, _ = writer.Write([]byte(`{"responseCode":0,"data":{"hash":"tx-123","fromAccountId":"buyer@bkrt","toAccountId":"oursell@bkrt","currency":"USD","amount":10.25,"description":"wallet top-up"}}`))
	}))
	defer server.Close()

	client := &bakongClient{apiToken: "test-token", baseURL: server.URL, httpClient: server.Client()}
	payment, err := client.CheckTransaction(context.Background(), "abc")
	if err != nil {
		t.Fatalf("CheckTransaction() error = %v", err)
	}
	if payment.Hash != "tx-123" || payment.ToAccount != "oursell@bkrt" || payment.Amount != 10.25 {
		t.Fatalf("payment = %+v", payment)
	}
}

func TestBakongClientMapsMissingTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"responseCode":1,"errorCode":1,"responseMessage":"transaction not found"}`))
	}))
	defer server.Close()

	client := &bakongClient{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.CheckTransaction(context.Background(), "missing")
	if !errors.Is(err, ErrPaymentNotFound) {
		t.Fatalf("CheckTransaction() error = %v, want ErrPaymentNotFound", err)
	}
}

func TestNewProviderRequiresCredentials(t *testing.T) {
	if provider := NewProvider(config.Config{}); provider != nil {
		t.Fatal("NewProvider() should disable Bakong without credentials")
	}
}
