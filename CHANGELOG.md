# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-08-20

### Added

- **Webhook receiver** for OpenProject events with HTTP POST endpoint at `/webhook`
- **Telegram Bot API integration** for forwarding notifications (zero external dependencies)
- **Work Package notifications** with subject, type, status, priority, author, assignee, project, and direct link
- **Time Entry notifications** with user, hours, date, activity, and project details
- **Project notifications** for project creation and updates
- **Membership notifications** with user, project, and role information
- **Generic fallback handler** for unknown/future OpenProject event types
- **HMAC-SHA1 signature verification** for webhook security (optional)
- **HTML formatted messages** with emoji indicators for Telegram
- **Health check endpoint** at `/health`
- **Graceful shutdown** with OS signal handling
- **Request logging middleware**
- **Docker support** with multi-stage build (Go builder to Alpine runtime)
- **Docker Compose** configuration for easy deployment
- **Environment-based configuration** (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID, WEBHOOK_SECRET, OPENPROJECT_URL, PORT)
- **GitHub Actions CI** workflows: build with SonarQube, Go lint and test, issue labeler
