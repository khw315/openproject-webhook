# OpenProject → Telegram Webhook Forwarder

Service standalone Golang yang menerima webhook dari [OpenProject](https://www.openproject.org/) dan meneruskan notifikasi ke [Telegram](https://telegram.org/) secara real-time.

## ✨ Fitur

- 📋 Notifikasi **Work Package** (created, updated, deleted)
- 💬 Notifikasi **Work Package Comment / Activity** (komentar & aktivitas)
- ⏱️ Notifikasi **Time Entry** (created)
- 📁 Notifikasi **Project** (created, updated)
- 👥 Notifikasi **Membership** (created, updated)
- 🔔 Fallback handler untuk event yang belum dikenal
- 🔒 Verifikasi signature webhook (opsional)
- 🔗 Link langsung ke OpenProject
- 🐳 Docker ready (multi-stage build, < 15MB)
- 🚀 Zero dependency — hanya Go standard library

## 📋 Prasyarat

- [Go 1.23+](https://go.dev/dl/) (jika build tanpa Docker)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)
- Akses admin ke OpenProject instance
- Telegram Bot Token

## 🤖 Langkah 1: Buat Telegram Bot

1. Buka Telegram, cari **@BotFather**
2. Kirim `/newbot`
3. Ikuti instruksi, beri nama dan username untuk bot
4. Simpan **Bot Token** yang diberikan (format: `123456789:ABCdef...`)
5. Tambahkan bot ke **group chat** atau **channel** yang diinginkan
6. Dapatkan **Chat ID**:
   - Untuk chat pribadi: kirim pesan ke bot, lalu buka `https://api.telegram.org/bot<TOKEN>/getUpdates`
   - Untuk group: tambahkan bot ke group, kirim pesan, lalu cek `getUpdates`
   - Chat ID group biasanya negatif, contoh: `-1001234567890`

## ⚙️ Langkah 2: Konfigurasi

```bash
# Clone / copy project
cp .env.example .env
```

Edit file `.env`:

```env
# [WAJIB] Token dari BotFather
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz

# [WAJIB] Chat ID tujuan
TELEGRAM_CHAT_ID=-1001234567890

# [OPSIONAL] Secret untuk verifikasi signature
WEBHOOK_SECRET=rahasia-webhook-saya

# [OPSIONAL] URL OpenProject untuk generate link
OPENPROJECT_URL=https://openproject.perusahaan.com

# [OPSIONAL] Port (default: 8080)
PORT=8080
```

## 🐳 Langkah 3: Deploy dengan Docker

### Opsi A: Docker Compose (Rekomendasi)

```bash
# Build dan jalankan
docker compose up -d

# Lihat logs
docker compose logs -f

# Stop
docker compose down
```

### Opsi B: Docker Manual

```bash
# Build image
docker build -t op-telegram-webhook .

# Jalankan container
docker run -d \
  --name op-telegram-webhook \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file .env \
  op-telegram-webhook
```

### Opsi C: Tanpa Docker

```bash
go build -o webhook-forwarder .
./webhook-forwarder
```

## 🔗 Langkah 4: Konfigurasi Webhook di OpenProject

1. Login ke OpenProject sebagai **Administrator**
2. Buka **Administration → API and webhooks**
3. Klik **+ Webhook**
4. Isi form:
   - **Name**: `Telegram Notifier` (atau nama lain)
   - **Payload URL**: `http://<server-ip>:8080/webhook`
   - **Signature secret**: isi sama dengan `WEBHOOK_SECRET` di `.env` (opsional)
   - **Enabled**: ✅ centang
5. Pilih **Events** yang ingin di-notifikasi:
   - ✅ Work package: created
   - ✅ Work package: updated
   - ✅ Time entry: created
   - ✅ Project: created
   - ✅ Project: updated
   - ✅ Membership: created
   - ✅ Membership: updated
   - (atau semua event)
6. Pilih **Projects**: All projects (atau pilih project tertentu)
7. Klik **Create**

> ⚠️ **Penting**: Pastikan di **System settings → Notifications**, opsi "Work package added" dan "Work package updated" sudah aktif, agar webhook bisa terpicu.

## 🧪 Test Manual

Kirim sample webhook payload untuk testing:

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "action": "work_package:created",
    "work_package": {
      "_type": "WorkPackage",
      "id": 42,
      "subject": "Implementasi fitur login",
      "description": { "format": "markdown", "raw": "Buat halaman login dengan OAuth2", "html": "" },
      "percentageDone": 0,
      "createdAt": "2026-08-20T10:00:00Z",
      "updatedAt": "2026-08-20T10:00:00Z",
      "_links": {
        "self": { "href": "/api/v3/work_packages/42", "title": "Implementasi fitur login" },
        "project": { "href": "/api/v3/projects/1", "title": "Website Redesign" },
        "type": { "href": "/api/v3/types/1", "title": "Task" },
        "status": { "href": "/api/v3/statuses/1", "title": "New" },
        "priority": { "href": "/api/v3/priorities/2", "title": "High" },
        "author": { "href": "/api/v3/users/1", "title": "Ahmad Faiz" },
        "assignee": { "href": "/api/v3/users/2", "title": "Budi Santoso" }
      }
    }
  }'
```

Health check:

```bash
curl http://localhost:8080/health
# {"status":"healthy"}
```

## 📱 Contoh Notifikasi Telegram

```
✨ Work Package Created
━━━━━━━━━━━━━━━━━━━━━
📌 Subject: Implementasi fitur login
🏷️ Type: Task
📊 Status: New
⚡ Priority: High
👤 Author: Ahmad Faiz
👷 Assignee: Budi Santoso
📁 Project: Website Redesign

📝 Buat halaman login dengan OAuth2

🔗 Open in OpenProject
```

## 🏗️ Arsitektur

```
┌──────────────┐      POST /webhook      ┌────────────────────┐      sendMessage      ┌──────────┐
│  OpenProject │ ──────────────────────▶ │  Go Webhook Server │ ────────────────────▶ │ Telegram │
│   Instance   │                         │     (port 8080)    │                       │ Bot API  │
└──────────────┘                         └────────────────────┘                       └──────────┘
```

## 📁 Struktur Project

```
openproject-telegram-webhook/
├── main.go              # Entry point, HTTP server, graceful shutdown
├── handler.go           # Webhook handler, signature verification
├── telegram.go          # Telegram Bot API client (zero dependency)
├── models.go            # OpenProject payload data structures
├── formatter.go         # Message formatting per event type
├── config.go            # Environment variable configuration
├── go.mod               # Go module definition
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose deployment
├── .env.example         # Environment variable template
└── README.md            # Dokumentasi ini
```

## 📝 Lisensi

MIT License
