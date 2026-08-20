package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// WebhookHandler handles incoming webhook requests from OpenProject.
type WebhookHandler struct {
	config   *Config
	telegram *TelegramClient
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(cfg *Config, tg *TelegramClient) *WebhookHandler {
	return &WebhookHandler{
		config:   cfg,
		telegram: tg,
	}
}

// ServeHTTP implements http.Handler for the webhook endpoint.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body.
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
	if err != nil {
		log.Printf("[webhook] failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature if a secret is configured.
	if h.config.WebhookSecret != "" {
		if !h.verifySignature(r, body) {
			log.Printf("[webhook] signature verification failed")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse the webhook payload.
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[webhook] failed to parse JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.Action == "" {
		log.Printf("[webhook] missing action field")
		http.Error(w, "Missing action field", http.StatusBadRequest)
		return
	}

	log.Printf("[webhook] received event: %s", payload.Action)

	// Format the message.
	message := FormatMessage(payload.Action, &payload, h.config.OpenProjectURL)

	// Send to Telegram.
	if err := h.telegram.SendMessage(message); err != nil {
		log.Printf("[webhook] failed to send telegram message: %v", err)
		http.Error(w, "Failed to forward to Telegram", http.StatusInternalServerError)
		return
	}

	log.Printf("[webhook] successfully forwarded event: %s", payload.Action)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

// verifySignature checks the X-OP-Signature header against the request body
// using HMAC-SHA1 with the configured webhook secret.
func (h *WebhookHandler) verifySignature(r *http.Request, body []byte) bool {
	signature := r.Header.Get("X-OP-Signature")
	if signature == "" {
		return false
	}

	// OpenProject sends the signature as "sha1=<hex>".
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha1" {
		return false
	}

	expectedMAC, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	mac := hmac.New(sha1.New, []byte(h.config.WebhookSecret))
	mac.Write(body)
	actualMAC := mac.Sum(nil)

	return hmac.Equal(actualMAC, expectedMAC)
}

// HealthHandler responds to health check requests.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
}
