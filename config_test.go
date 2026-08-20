package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	os.Setenv("OPENPROJECT_URL", "https://openproject.example.com/")
	os.Setenv("PORT", "9000")
	defer func() {
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("TELEGRAM_CHAT_ID")
		os.Unsetenv("OPENPROJECT_URL")
		os.Unsetenv("PORT")
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
}

func TestLoadConfigMissing(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("TELEGRAM_CHAT_ID")

	_, err := LoadConfig()
	if err == nil {
		t.Errorf("expected error when token is missing")
	}
}
