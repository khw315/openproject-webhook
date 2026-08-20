package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func generateSignature(secret string, body []byte) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(body)
	return fmt.Sprintf("sha1=%s", hex.EncodeToString(mac.Sum(nil)))
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	expected := "{\"status\":\"healthy\"}"
	if rec.Body.String() != expected {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestWebhookHandler(t *testing.T) {
	var telegramShouldFail bool
	tgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if telegramShouldFail {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("{\"ok\":false}"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"ok\":true}"))
	}))
	defer tgServer.Close()

	cfg := &Config{
		TelegramBotToken: "test",
		TelegramChatID:   "123",
		WebhookSecret:    "mysecret",
		OpenProjectURL:   "https://op.example.com",
		Port:             "8080",
	}
	tgClient := NewTelegramClient(cfg.TelegramBotToken, cfg.TelegramChatID)
	tgClient.apiBaseURL = tgServer.URL
	tgClient.httpClient = tgServer.Client()

	handler := NewWebhookHandler(cfg, tgClient)

	// 1. Method Not Allowed
	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// 2. Invalid Signature (secret set, no header)
	body := []byte("{\"action\":\"work_package:created\"}")
	req = httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}

	// 3. Invalid JSON
	invalidJSON := []byte("{not json}")
	req = httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(invalidJSON))
	req.Header.Set("X-OP-Signature", generateSignature("mysecret", invalidJSON))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid json, got %d", rec.Code)
	}

	// 4. Missing action field
	emptyAction := []byte("{}")
	req = httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(emptyAction))
	req.Header.Set("X-OP-Signature", generateSignature("mysecret", emptyAction))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing action, got %d", rec.Code)
	}

	// 5. Telegram send failure
	telegramShouldFail = true
	req = httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("X-OP-Signature", generateSignature("mysecret", body))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when telegram fails, got %d", rec.Code)
	}

	// 6. Success forward
	telegramShouldFail = false
	req = httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("X-OP-Signature", generateSignature("mysecret", body))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for successful webhook forward, got %d", rec.Code)
	}
	if rec.Body.String() != "{\"status\":\"ok\"}" {
		t.Errorf("unexpected response body: %s", rec.Body.String())
	}
}

func TestVerifySignature(t *testing.T) {
	cfg := &Config{WebhookSecret: "secretkey"}
	handler := NewWebhookHandler(cfg, nil)
	body := []byte("test payload")

	// 1. Missing header
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	if handler.verifySignature(req, body) {
		t.Error("expected false for missing header")
	}

	// 2. Bad prefix
	req.Header.Set("X-OP-Signature", "md5=12345")
	if handler.verifySignature(req, body) {
		t.Error("expected false for invalid prefix")
	}

	// 3. Bad hex
	req.Header.Set("X-OP-Signature", "sha1=not-hex-characters-zz")
	if handler.verifySignature(req, body) {
		t.Error("expected false for non-hex signature")
	}

	// 4. Mismatch signature
	req.Header.Set("X-OP-Signature", "sha1=00112233445566778899aabbccddeeff00112233")
	if handler.verifySignature(req, body) {
		t.Error("expected false for mismatched signature")
	}

	// 5. Valid signature
	validSig := generateSignature("secretkey", body)
	req.Header.Set("X-OP-Signature", validSig)
	if !handler.verifySignature(req, body) {
		t.Error("expected true for valid signature")
	}
}
