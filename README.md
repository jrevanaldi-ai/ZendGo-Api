# ZendGo - WhatsApp Unofficial API

## Apa Ini?

ZendGo itu API WhatsApp unofficial yang aku bikin pakai Go (Golang) dan library whatsmeow. Jadi intinya ini jembatan antara aplikasi kamu dengan WhatsApp Web, biar kamu bisa kirim pesan, gambar, video, dll lewat API REST.

## Kenapa Aku Bikin Ini?

Aku butuh API WhatsApp yang:
- Murah (gratis, nggak bayar per pesan kayak official API)
- Cepat dan ringan
- Bisa di-hosting sendiri
- Support multi-session (banyak nomor WhatsApp dalam satu server)

Nah, ZendGo ini jawabannya.

---

## Fitur Yang Ada

Ini yang udah aku implementasi dan udah aku test sendiri:

### Session & Authentication
- Create session (satu session = satu nomor WhatsApp)
- Pairing pakai QR Code
- Pairing pakai Pairing Code (8 digit)
- Multi-session support
- API Key per session

### Kirim Pesan
- Text message
- Image (dari URL)
- Video (dari URL)
- Document (dari URL)
- Audio (dari URL)
- Location (latitude/longitude)
- Contact (VCard)

### Group Management
- List groups
- Create group
- Add/remove participants
- Update group info

### Webhook
- Notifikasi pesan masuk real-time
- Callback ke URL kamu

### Database
- PostgreSQL untuk storage
- Auto migrations
- Session persistence

---

## Cara Install

### 1. Install Go

Kamu butuh Go 1.25.0 atau lebih baru. Cek dulu:

```bash
go version
```

Kalau belum ada, install dulu.

### 2. Install PostgreSQL

ZendGo pakai PostgreSQL untuk nyimpen session dan message.

```bash
# Di Termux
pkg install postgresql

# Start PostgreSQL
pg_ctl -D $PREFIX/var/lib/postgresql start

# Buat user dan database
createuser -s postgres
createdb -U postgres zendgo
```

### 3. Clone atau Download Project

```bash
cd /path/to/project
```

### 4. Install Dependencies

```bash
go mod tidy
```

### 5. Build

```bash
go build -o zendgo ./cmd/server
```

### 6. Setup Environment

Copy file `.env.example` jadi `.env`:

```bash
cp .env.example .env
```

Edit `.env` sesuai kebutuhan:

```
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=zendgo
DB_SSLMODE=disable

WHATSAPP_LOG_LEVEL=info
```

### 7. Jalankan

```bash
./zendgo
```

Server akan jalan di `http://localhost:8080`

---

## Cara Pakai

### Step 1: Buat Session

Session itu kayak "koneksi" antara ZendGo dengan nomor WhatsApp kamu. Satu session = satu nomor.

```bash
curl -X POST http://localhost:8080/api/v1/sessions/new \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "628123456789",
    "webhook_url": "https://your-domain.com/webhook"
  }'
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "session-uuid-here",
    "phone": "628123456789",
    "status": "connecting",
    "api_key": "your-api-key-here"
  },
  "message": "Session created successfully. Scan QR code to connect."
}
```

**PENTING:** Simpan `session_id` dan `api_key` nya! Kamu butuh ini untuk request selanjutnya.

### Step 2: Pairing (Connect ke WhatsApp)

Ada 2 cara: QR Code atau Pairing Code.

#### Cara A: Pairing Code (Lebih Mudah)

```bash
curl -X POST "http://localhost:8080/api/v1/sessions/pair?id=SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "628123456789",
    "client_type": "chrome",
    "device_name": "Chrome (Linux)"
  }'
```

Response:

```json
{
  "success": true,
  "data": "ABCD1234",
  "message": "Enter this code on your WhatsApp app"
}
```

Nah, sekarang buka WhatsApp di HP kamu:
1. Settings → Linked Devices
2. Link a Device
3. Pilih "Use code instead" atau "Pair with code"
4. Masukkan code: `ABCD1234`

Done! Session status jadi `paired`.

#### Cara B: QR Code

```bash
curl "http://localhost:8080/api/v1/sessions/qr?id=SESSION_ID"
```

Response:

```json
{
  "success": true,
  "data": "1@AbCdEfGhIjKlMnOpQrStUvWxYz1234567890,base64qrcode..."
}
```

Scan QR code ini pakai WhatsApp:
1. Settings → Linked Devices
2. Link a Device
3. Scan QR code

### Step 3: Kirim Pesan

Setelah paired, kamu bisa kirim pesan pakai `api_key` yang tadi.

#### Kirim Text

```bash
curl -X POST http://localhost:8080/api/v1/messages/text \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "recipient": "628987654321",
    "message": "Halo dari ZendGo!"
  }'
```

#### Kirim Gambar

```bash
curl -X POST http://localhost:8080/api/v1/messages/image \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "recipient": "628987654321",
    "image_url": "https://example.com/image.jpg",
    "caption": "Ini gambarnya"
  }'
```

#### Kirim Video

```bash
curl -X POST http://localhost:8080/api/v1/messages/video \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "recipient": "628987654321",
    "video_url": "https://example.com/video.mp4",
    "caption": "Ini videonya"
  }'
```

#### Kirim Document

```bash
curl -X POST http://localhost:8080/api/v1/messages/document \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "recipient": "628987654321",
    "document_url": "https://example.com/file.pdf",
    "file_name": "document.pdf",
    "caption": "Ini dokumennya"
  }'
```

#### Kirim Audio

```bash
curl -X POST http://localhost:8080/api/v1/messages/audio \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "recipient": "628987654321",
    "audio_url": "https://example.com/audio.mp3"
  }'
```

#### Kirim Lokasi

```bash
curl -X POST http://localhost:8080/api/v1/messages/location \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "recipient": "628987654321",
    "latitude": -6.2088,
    "longitude": 106.8456,
    "name": "Jakarta",
    "address": "Jakarta, Indonesia"
  }'
```

#### Kirim Contact

```bash
curl -X POST http://localhost:8080/api/v1/messages/contact \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "recipient": "628987654321",
    "display_name": "John Doe",
    "phone_number": "628123456789",
    "organization": "Acme Corp"
  }'
```

### Step 4: Manage Group

#### List Groups

```bash
curl -X GET http://localhost:8080/api/v1/groups/list \
  -H "X-API-Key: YOUR_API_KEY"
```

#### Create Group

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "name": "Grup Baru",
    "participants": ["628111111111", "628222222222"]
  }'
```

---

## Workflow & Alur Kerja

Gini alur kerjanya dari awal sampai bisa kirim pesan:

```
┌─────────────┐
│   Kamu      │
│  (Client)   │
└──────┬──────┘
       │
       │ 1. POST /sessions/new
       ▼
┌─────────────────┐
│   ZendGo API    │
│                 │
│  - Create       │
│    session di   │
│    database     │
│  - Generate     │
│    API Key      │
└──────┬──────────┘
       │
       │ 2. Return session_id & api_key
       ▼
┌─────────────┐
│   Kamu      │
└──────┬──────┘
       │
       │ 3. POST /sessions/pair
       ▼
┌─────────────────┐
│   ZendGo API    │
│                 │
│  - Connect ke   │
│    WhatsApp     │
│    Web          │
│  - Generate     │
│    pairing code │
└──────┬──────────┘
       │
       │ 4. Return pairing code
       ▼
┌─────────────┐
│   Kamu      │
└──────┬──────┘
       │
       │ 5. Masukkan code di WhatsApp
       ▼
┌─────────────────┐
│   WhatsApp      │
│   Web Server    │
│                 │
│  - Validate     │
│  - Link device  │
└──────┬──────────┘
       │
       │ 6. Pair success
       ▼
┌─────────────────┐
│   ZendGo API    │
│                 │
│  - Update       │
│    session      │
│    status =     │
│    "paired"     │
└──────┬──────────┘
       │
       │ 7. Session ready!
       ▼
┌─────────────┐
│   Kamu      │
└──────┬──────┘
       │
       │ 8. POST /messages/text (dengan api_key)
       ▼
┌─────────────────┐
│   ZendGo API    │
│                 │
│  - Validate     │
│    API Key      │
│  - Send via     │
│    WhatsApp Web │
└──────┬──────────┘
       │
       │ 9. Pesan terkirim!
       ▼
┌─────────────────┐
│   Recipient     │
│   WhatsApp      │
└─────────────────┘
```

### Alur Pesan Masuk (Webhook)

```
┌─────────────┐
│  Pengirim   │
└──────┬──────┘
       │
       │ 1. Kirim pesan ke nomor WhatsApp kamu
       ▼
┌─────────────────┐
│   WhatsApp      │
│   Web Server    │
└──────┬──────────┘
       │
       │ 2. Push event ke ZendGo
       ▼
┌─────────────────┐
│   ZendGo API    │
│                 │
│  - Receive      │
│    event        │
│  - Save to DB   │
│  - Forward to   │
│    webhook URL  │
└──────┬──────────┘
       │
       │ 3. HTTP POST ke webhook_url kamu
       ▼
┌─────────────────┐
│   Your Server   │
│   (Webhook)     │
└─────────────────┘
```

---

## Struktur Project

Gini struktur folder ZendGo:

```
ZendGo/
├── cmd/
│   └── server/
│       └── main.go          # Entry point aplikasi
├── internal/
│   ├── config/              # Configuration loader
│   ├── handler/             # HTTP handlers (controllers)
│   ├── service/             # Business logic
│   ├── repository/          # Database access layer
│   └── middleware/          # Auth, logging, CORS
├── pkg/
│   └── whatsapp/            # Whatsmeow wrapper
│       ├── client.go        # WhatsApp client setup
│       ├── message.go       # Send message functions
│       └── session.go       # Group & session functions
├── models/                  # Data models & migrations
├── routes/                  # API route definitions
├── debugging/
│   └── log.txt              # Log file
├── .env.example             # Environment template
├── go.mod                   # Go dependencies
└── zendgo                   # Binary executable
```

### Penjelasan Tiap Folder

**cmd/server/** - Ini entry point. Kalau kamu run `./zendgo`, file `main.go` di sini yang jalan pertama kali.

**internal/config/** - Load konfigurasi dari environment variables (.env file).

**internal/handler/** - HTTP handlers. Ini yang handle request dari client (curl, Postman, aplikasi kamu). Ada `session_handler.go` untuk handle session dan `message_handler.go` untuk handle pesan.

**internal/service/** - Business logic layer. Di sini logika utamanya, kayak gimana cara create session, gimana cara kirim pesan, dll.

**internal/repository/** - Database layer. Semua query ke PostgreSQL ada di sini.

**internal/middleware/** - Middleware kayak authentication (cek API key), logging, CORS.

**pkg/whatsapp/** - Wrapper untuk whatsmeow library. Ini jembatan antara ZendGo sama WhatsApp Web.

**models/** - Data models (struct Go) dan migration SQL untuk database.

**routes/** - Definisi semua endpoint API. Di sini kamu bisa lihat semua URL yang tersedia.

**debugging/** - Folder untuk log file.

---

## API Endpoints

Semua endpoint yang tersedia:

### Public Endpoints (No Auth)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/health` | Cek apakah API jalan |
| POST | `/api/v1/sessions/new` | Buat session baru |
| GET | `/api/v1/sessions/qr?id=ID` | Ambil QR code |
| POST | `/api/v1/sessions/pair?id=ID` | Pairing pakai code |
| GET | `/api/v1/sessions` | List semua session |
| DELETE | `/api/v1/sessions/delete?id=ID` | Hapus session |

### Protected Endpoints (Butuh API Key)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/v1/messages/text` | Kirim text |
| POST | `/api/v1/messages/image` | Kirim gambar |
| POST | `/api/v1/messages/video` | Kirim video |
| POST | `/api/v1/messages/document` | Kirim dokumen |
| POST | `/api/v1/messages/audio` | Kirim audio |
| POST | `/api/v1/messages/location` | Kirim lokasi |
| POST | `/api/v1/messages/contact` | Kirim contact |
| POST | `/api/v1/messages/cta-button` | Kirim CTA button (text format) |
| POST | `/api/v1/groups` | Buat group |
| GET | `/api/v1/groups/list` | List groups |

---

## Format Request & Response

### Health Check

```bash
GET /health
```

Response:
```json
{
  "success": true,
  "message": "ZendGo API is running"
}
```

### Create Session

```bash
POST /api/v1/sessions/new
Content-Type: application/json

{
  "phone": "628123456789",
  "webhook_url": "https://your-domain.com/webhook"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid-session",
    "phone": "628123456789",
    "status": "connecting",
    "webhook_url": "https://your-domain.com/webhook",
    "api_key": "your-api-key"
  },
  "message": "Session created successfully. Scan QR code to connect."
}
```

### Pairing

```bash
POST /api/v1/sessions/pair?id=SESSION_ID
Content-Type: application/json

{
  "phone": "628123456789",
  "client_type": "chrome",
  "device_name": "Chrome (Linux)"
}
```

Response:
```json
{
  "success": true,
  "data": "ABCD1234",
  "message": "Enter this code on your WhatsApp app"
}
```

### Send Message

```bash
POST /api/v1/messages/text
X-API-Key: YOUR_API_KEY
Content-Type: application/json

{
  "recipient": "628987654321",
  "message": "Halo!"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid-message",
    "session_id": "uuid-session",
    "recipient": "628987654321",
    "message_type": "text",
    "content": "Halo!",
    "status": "sent",
    "wa_message_id": "3EB0..."
  },
  "message": "Message sent successfully"
}
```

### Error Response

```json
{
  "success": false,
  "message": "Error message here"
}
```

---

## Webhook

Kalau ada pesan masuk, ZendGo akan HTTP POST ke `webhook_url` yang kamu set saat create session.

Format webhook:

```json
{
  "session_id": "uuid-session",
  "sender": "628123456789@s.whatsapp.net",
  "message_type": "text",
  "content": "Halo!",
  "caption": "",
  "timestamp": 1234567890,
  "is_group": false,
  "from": "628123456789@s.whatsapp.net",
  "push_name": "John Doe"
}
```

Kamu perlu setup server untuk terima webhook ini. Contoh pakai Node.js:

```javascript
app.post('/webhook', (req, res) => {
  const { sender, content, message_type } = req.body;
  console.log(`Pesan dari ${sender}: ${content}`);
  res.json({ success: true });
});
```

---

## Catatan Penting

### 1. Nomor WhatsApp Harus Online

Nomor WhatsApp yang kamu pairing harus tetap online dan terconnect ke internet. Kalau logout atau disconnect, session status jadi `disconnected` dan perlu pairing ulang.

### 2. Rate Limit WhatsApp

WhatsApp punya rate limit. Jangan spam kirim pesan ke banyak nomor dalam waktu singkat, bisa banned!

### 3. Session Expires

Session bisa expired kalau:
- Logout dari WhatsApp
- Lama tidak aktif
- WhatsApp Web disconnect

Kalau expired, perlu create session baru dan pairing ulang.

### 4. Format Nomor

Format nomor yang benar:
- ✅ `628123456789` (international, tanpa + atau 00)
- ✅ `628123456789@s.whatsapp.net` (dengan JID)
- ❌ `+628123456789` (dengan +)
- ❌ `08123456789` (dengan 0 di depan)

### 5. Multi-Session

Kamu bisa punya banyak session dengan nomor berbeda. Tiap session punya `api_key` berbeda, jadi bisa dipake untuk client berbeda.

Contoh:
- Session 1: Nomor Marketing (API Key: abc123)
- Session 2: Nomor Support (API Key: xyz789)
- Session 3: Nomor Sales (API Key: def456)

---

## Troubleshooting

### "Session not found"

Ini terjadi kalau server restart tapi session masih di database. Solusinya:
1. Delete session lama: `DELETE /api/v1/sessions/delete?id=ID`
2. Create session baru: `POST /api/v1/sessions/new`
3. Pairing ulang

### "Session not paired"

Session belum pairing. Pairing dulu pakai QR code atau pairing code.

### "Failed to send message"

Cek:
- Session status harus `paired`
- Nomor recipient format benar
- API key valid

### "Websocket not connected"

WhatsApp Web belum connect. Tunggu beberapa detik setelah create session, baru pairing.

---

## License

MIT License - Bebas dipake untuk personal atau commercial project.

---

## Disclaimer

Ini unofficial API. Pakai dengan bijak dan tanggung jawab sendiri. Jangan pakai untuk spam atau hal ilegal. Pastikan comply dengan Terms of Service WhatsApp.

---

## Contact

Kalau ada pertanyaan atau issue, silakan buka issue di repository atau hubungi aku langsung.

---

Dibuat dengan ❤️ menggunakan Go dan whatsmeow.
