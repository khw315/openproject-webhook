package main

import (
	"fmt"
	"os"
)

// Config holds all configuration values for the webhook forwarder.
type Config struct {
	// TelegramBotToken is the Bot API token from @BotFather.
	TelegramBotToken string

	// TelegramChatID is the target chat/group/channel ID.
	TelegramChatID string

	// WebhookSecret is the optional shared secret for signature verification.
	WebhookSecret string

	// OpenProjectURL is the base URL of the OpenProject instance (for generating links).
	OpenProjectURL string

	// Port is the HTTP server listen port.
	Port string
}

// LoadConfig reads configuration from environment variables and validates required fields.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		WebhookSecret:    os.Getenv("WEBHOOK_SECRET"),
		OpenProjectURL:   os.Getenv("OPENPROJECT_URL"),
		Port:             os.Getenv("PORT"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.TelegramChatID == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Strip trailing slash from OpenProject URL if present.
	if len(cfg.OpenProjectURL) > 0 && cfg.OpenProjectURL[len(cfg.OpenProjectURL)-1] == '/' {
		cfg.OpenProjectURL = cfg.OpenProjectURL[:len(cfg.OpenProjectURL)-1]
	}

	return cfg, nil
}
