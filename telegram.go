package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TelegramClient sends messages to Telegram via the Bot API.
type TelegramClient struct {
	botToken   string
	chatID     string
	httpClient *http.Client
}

// NewTelegramClient creates a new Telegram client with the given bot token and chat ID.
func NewTelegramClient(botToken, chatID string) *TelegramClient {
	return &TelegramClient{
		botToken: botToken,
		chatID:   chatID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// sendMessageRequest is the JSON body for the Telegram sendMessage API.
type sendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

// sendMessageResponse is the JSON response from the Telegram sendMessage API.
type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
}

// SendMessage sends a formatted message to the configured Telegram chat.
// It uses HTML parse mode. Retries once on failure.
func (tc *TelegramClient) SendMessage(text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tc.botToken)

	payload := sendMessageRequest{
		ChatID:                tc.chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	// Try sending, retry once on failure.
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			log.Printf("[telegram] retrying send (attempt %d)...", attempt+1)
			time.Sleep(2 * time.Second)
		}

		lastErr = tc.doSend(url, payload)
		if lastErr == nil {
			return nil
		}
		log.Printf("[telegram] send failed: %v", lastErr)
	}

	return fmt.Errorf("failed to send telegram message after 2 attempts: %w", lastErr)
}

// doSend performs the actual HTTP POST to Telegram.
func (tc *TelegramClient) doSend(url string, payload sendMessageRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result sendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	return nil
}
