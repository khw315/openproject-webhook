package main

import (
	"os"
	"testing"
)

func TestLoadConfigSuccess(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	os.Setenv("OPENPROJECT_URL", "https://openproject.example.com/")
	os.Setenv("PORT", "9000")
	os.Setenv("WEBHOOK_SECRET", "secret123")
	defer func() {
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("TELEGRAM_CHAT_ID")
		os.Unsetenv("OPENPROJECT_URL")
		os.Unsetenv("PORT")
		os.Unsetenv("WEBHOOK_SECRET")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.TelegramBotToken != "test-token" {
		t.Errorf("expected test-token, got %s", cfg.TelegramBotToken)
	}
	if cfg.OpenProjectURL != "https://openproject.example.com" {
		t.Errorf("expected trailing slash removed, got %s", cfg.OpenProjectURL)
	}
	if cfg.Port != "9000" {
		t.Errorf("expected 9000, got %s", cfg.Port)
	}
	if cfg.WebhookSecret != "secret123" {
		t.Errorf("expected secret123, got %s", cfg.WebhookSecret)
	}
}

func TestLoadConfigDefaultPortAndNoTrailingSlash(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	os.Setenv("OPENPROJECT_URL", "https://openproject.example.com")
	os.Unsetenv("PORT")
	defer func() {
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("TELEGRAM_CHAT_ID")
		os.Unsetenv("OPENPROJECT_URL")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
}

func TestLoadConfigMissingToken(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	_, err := LoadConfig()
	if err == nil {
		t.Errorf("expected error when token is missing")
	}
}

func TestLoadConfigMissingChatID(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Unsetenv("TELEGRAM_CHAT_ID")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")

	_, err := LoadConfig()
	if err == nil {
		t.Errorf("expected error when chat ID is missing")
	}
}
