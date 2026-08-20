# OpenProject Telegram Webhook Forwarder

A lightweight, standalone webhook service written in Go that receives webhooks from [OpenProject](https://www.openproject.org/) and forwards real-time notifications to [Telegram](https://telegram.org/).

## Features

- **Work Package Tracking**: Notifications for creations, updates, and deletions with full details (type, status, priority, author, assignee, progress, and direct links).
- **Work Package Comments**: Real-time notifications for work package comments and journal activity entries.
- **Time Tracking**: Forward time entry logs including hours, spent dates, activities, and user notes.
- **Project & Membership**: Notifications for project lifecycle events and team membership/role changes.
- **Generic Fallback**: Safely handles unknown or newly introduced OpenProject event types.
- **Signature Verification**: HMAC-SHA1 signature verification (`X-OP-Signature`) for webhook security.
- **Zero External Dependencies**: Built entirely with the Go standard library.
- **Ultra Lightweight**: Docker image size under 15MB using multi-stage build.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) (or [Go 1.23+](https://go.dev/dl/) for bare-metal builds)
- Administrator access to an OpenProject instance
- A Telegram bot token and target chat ID

## Quick Start

### 1. Create a Telegram Bot

1. Open Telegram and search for [@BotFather](https://t.me/BotFather).
2. Send `/newbot` and follow the prompts to obtain your **Bot Token** (`123456789:ABCdef...`).
3. Create or open your target group/channel and add the bot as an administrator (or start a direct chat with `/start`).
4. Retrieve your **Chat ID**:
   - For private chat: Check your ID via [@userinfobot](https://t.me/userinfobot) or `https://api.telegram.org/bot<TOKEN>/getUpdates`.
   - For groups/channels: Send a message in the group and check `getUpdates` (group IDs usually start with a negative number, e.g. `-1001234567890`).

### 2. Configure Environment

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```env
# Required
TELEGRAM_BOT_TOKEN=123456789:AAH_0NvRbdSLVM91I7XrVmTMJhIRnc5AaV8
TELEGRAM_CHAT_ID=-1001234567890

# Optional
WEBHOOK_SECRET=your-webhook-secret
OPENPROJECT_URL=https://openproject.example.com
PORT=8080
```

### 3. Deploy

#### Option A: Docker Compose (Recommended)

```bash
docker compose up -d
```

View container logs:

```bash
docker compose logs -f
```

#### Option B: Docker Container

```bash
docker build -t openproject-webhook .
docker run -d \
  --name openproject-webhook \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file .env \
  openproject-webhook
```

#### Option C: Native Go Binary

```bash
go build -o webhook-forwarder .
./webhook-forwarder
```

## OpenProject Webhook Setup

1. Log in to your OpenProject instance as an **Administrator**.
2. Navigate to **Administration** > **API and webhooks** > **Webhooks**.
3. Click **+ Webhook** and fill in the configuration:
   - **Name**: `Telegram Notifications`
   - **Payload URL**: `http://<your-server-host>:8080/webhook`
   - **Signature secret**: Must match `WEBHOOK_SECRET` in `.env` (leave blank if not using signature verification)
   - **Enabled**: Checked
4. Select the **Events** you want to receive:
   - Work package (`created`, `updated`, `deleted`)
   - Work package comments / activities
   - Time entry (`created`)
   - Project (`created`, `updated`)
   - Membership (`created`, `updated`)
5. Select target **Projects** (or choose *All projects*).
6. Click **Save**.

> [!IMPORTANT]
> If OpenProject and the webhook forwarder run in separate Docker containers, ensure OpenProject can reach the forwarder host (use the container service name on a shared Docker network or your server's LAN/public IP, not `localhost`).

> [!TIP]
> Ensure notifications are enabled under **Administration** > **System settings** > **Notifications** in OpenProject so that work package events are triggered properly.

## Notification Examples

### Work Package Update

```text
Work Package Updated
━━━━━━━━━━━━━━━━━━━━━
Subject: Fix OAuth2 authentication redirect
Type: Bug
Status: In Progress
Priority: High
Author: Alice Smith
Assignee: Bob Johnson
Project: Core Platform
Progress: 75%
Due: 2026-08-25

Resolved redirect loop when callback URL contains parameters.

Open in OpenProject
```

### Work Package Comment

```text
Work Package Comment
━━━━━━━━━━━━━━━━━━━━━
Subject: Fix OAuth2 authentication redirect
Project: Core Platform
Author: Bob Johnson

"Fixed in commit 345a7c3. Ready for staging review."

Open in OpenProject
```

### Time Entry Log

```text
Time Entry Created
━━━━━━━━━━━━━━━━━━━━━
User: Alice Smith
Hours: 3.5
Date: 2026-08-20
Activity: Development
Work Package: Fix OAuth2 authentication redirect
Project: Core Platform

Code review and integration tests.
```

## Testing & Health Check

### Health Check

Verify that the service is running:

```bash
curl http://localhost:8080/health
# {"status":"healthy"}
```

### Manual Webhook Simulation

Simulate a Work Package creation event:

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "action": "work_package:created",
    "work_package": {
      "id": 101,
      "subject": "Test Telegram Notification",
      "_links": {
        "type": {"title": "Task"},
        "status": {"title": "New"},
        "priority": {"title": "Normal"},
        "project": {"title": "Demo Project"},
        "author": {"title": "Admin User"}
      }
    }
  }'
```

## Architecture

```
┌────────────────────┐     POST /webhook     ┌────────────────────────┐     Telegram Bot API     ┌──────────────────┐
│                    │ ────────────────────▶ │                        │ ───────────────────────▶ │                  │
│ OpenProject Server │   (JSON + HMAC-SHA1)  │ Webhook Forwarder (Go) │      (HTML Message)      │  Telegram Chat   │
│                    │                       │      Port: 8080        │                          │ (Group / Channel)│
└────────────────────┘                       └────────────────────────┘                          └──────────────────┘
```

## Project Structure

```
.
├── main.go            # Application entry point, HTTP routing & graceful shutdown
├── handler.go         # Webhook receiver & HMAC signature validation
├── formatter.go       # Telegram message formatters for OpenProject resources
├── models.go          # HAL+JSON and OpenProject webhook data models
├── telegram.go        # Telegram Bot API client with retry mechanism
├── config.go          # Environment configuration loader
├── Dockerfile         # Multi-stage container build
├── docker-compose.yml # Compose service configuration
└── .env.example       # Environment template
```
