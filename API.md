# Genfity WA Support API (Non-Provider Endpoints)

Dokumen ini hanya berisi endpoint milik service `genfity-wa-support`.

Tidak mencakup endpoint gateway ke provider (`/wa/*`) atau endpoint milik `genfity-wa`.

## Base URL

- Local: `http://localhost:8070` (atau sesuai `PORT`)

## Authentication

### 1) Internal API (antar service)
Gunakan salah satu header:
- `x-internal-api-key: <key>`
- `Authorization: Bearer <key>`

Format env (`INTERNAL_API_KEYS`):
- Scoped key (recommended): `service-name:key-value`
- Global key (legacy): `key-value`
- Contoh: `genfity-app:key123,govconnect:key456,super_admin_key`

### 2) Public Customer API
- `x-api-key: <customer_api_key>`

## System Endpoints

### `GET /`
Home/service info.

### `GET /health`
Health check.

### `HEAD /health`
Health check untuk probe/container.

---

## Internal Endpoints (`/internal/*`)

### `GET /internal/me`
Cek key internal yang dipakai:
- mode: `scoped` atau `global`
- source_service (kalau scoped)

**Response 200**
```json
{
  "auth": {
    "mode": "scoped",
    "source_service": "genfity-app"
  }
}
```

### `GET /internal/users?source=<service>&provider=genfity-wa&page=1&limit=20`
List user + ringkasan subscription + jumlah session.

Catatan:
- Jika key scoped, filter `source` otomatis dipaksa ke service pemilik key.
- `limit` max 100.

### `POST /internal/users`
Create/upsert user subscription.

**Body**
```json
{
  "user_id": "usr_001",
  "source": "genfity-app",
  "expires_at": "2026-12-31T23:59:59Z",
  "max_sessions": 3,
  "max_messages": 10000,
  "provider": "genfity-wa",
  "created_by": "order-service"
}
```

**Response 200**
- `api_key` hanya muncul ketika user baru dibuat.

### `PUT /internal/users/:user_id`
Update source/subscription user.

### `GET /internal/users/:user_id/apikey`
Metadata API key user (plaintext key tidak bisa dibaca ulang).

### `POST /internal/users/:user_id/apikey/rotate`
Rotate customer API key dan mengembalikan plaintext key baru.

---

## Public Customer Endpoints (`/v1/*`)

Semua endpoint berikut butuh `x-api-key`.

### `GET /v1/me`
Get info user + subscription aktif.

### `GET /v1/sessions`
List semua session milik user.

### `POST /v1/sessions`
Buat session baru.

**Body (umum)**
```json
{
  "session_name": "marketing-1",
  "webhook_url": "https://example.com/webhook",
  "events": "Message,Connected,Disconnected,QR",
  "expiration_sec": 0,
  "auto_connect": true,
  "auto_read_enabled": false,
  "typing_enabled": true,
  "history": 0
}
```

### `PUT /v1/sessions/:session_id`
Update konfigurasi session.

### `DELETE /v1/sessions/:session_id`
Hapus session.

### `GET /v1/sessions/:session_id/settings`
Get setting per session (`auto_read_enabled`, `typing_enabled`, `webhook_url`, message stats).

### `PUT /v1/sessions/:session_id/settings`
Update setting per session.

### `GET /v1/sessions/:session_id/contacts?sync=true|false`
List kontak per session.

Catatan:
- Default `sync=true` (auto sync dulu dari provider, lalu return cache DB lokal).
- `sync=false` hanya baca cache DB lokal.

### `POST /v1/sessions/:session_id/contacts/sync`
Paksa sync kontak dari provider ke DB lokal.

---

## Error Status (umum)

- `400` bad request/payload invalid
- `401` unauthorized (missing/invalid key)
- `403` forbidden (scope key, ownership, atau subscription)
- `404` resource tidak ditemukan
- `429` rate limit / spam block
- `500` internal server error
- `502` upstream/provider error

---

## Excluded from this document

- Semua endpoint `ANY /wa/*`
- Semua endpoint admin/provider yang langsung ke `genfity-wa`
