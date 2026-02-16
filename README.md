# Genfity WA Support (Normalized)

Service ini sekarang menjadi gateway WA mandiri dengan database sendiri, tanpa ketergantungan validasi subscription ke `genfity-app`.

## Scope Baru

- Menyimpan `user_id` + subscription WA di DB internal.
- Menyimpan session WA per user (`1 user_id` bisa punya banyak session).
- Menyimpan setting per session (`auto_read`, `typing`, `webhook`) dan statistik pesan per session.
- API internal (antar service) menggunakan `x-internal-api-key`.
- API customer/public menggunakan `x-api-key` per user.
- Request WA API tetap diproxy ke `genfity-wa`, termasuk sinkronisasi session.
- Fitur lama dihapus: campaign/blast, message history, event webhook storage, chat room/chat message persistence.

## Endpoint Utama

Dokumentasi API lengkap (khusus endpoint non-provider): lihat `API.md`.

### Public Customer
- `GET /v1/me`
- `GET /v1/sessions`
- `POST /v1/sessions`
- `PUT /v1/sessions/:session_id`
- `DELETE /v1/sessions/:session_id`
- `GET /v1/sessions/:session_id/settings`
- `PUT /v1/sessions/:session_id/settings`
- `GET /v1/sessions/:session_id/contacts`
- `POST /v1/sessions/:session_id/contacts/sync`

Catatan kontak:
- `GET /v1/sessions/:session_id/contacts` akan auto-sync dari `genfity-wa` secara default (`?sync=true`).
- Gunakan `?sync=false` jika ingin baca cache DB lokal saja.
- `genfity-wa` saat ini hanya expose `GET /user/contacts`, belum ada endpoint add contact manual.

### Internal Service-to-Service
- `GET /internal/me` (cek key ini scoped ke source apa atau global)
- `GET /internal/users?source=<service>&page=1&limit=20` (list user milik service tertentu)
- `POST /internal/users` (create/upsert user + subscription)
- `PUT /internal/users/:user_id` (update subscription)
- `GET /internal/users/:user_id/apikey` (metadata)
- `POST /internal/users/:user_id/apikey/rotate` (rotate dan return plaintext key baru)

Format key internal di `.env`:
- `INTERNAL_API_KEYS=service-a:keyA,service-b:keyB`
- Key scoped hanya boleh akses user dengan `source_service` yang sama.
- Format key lama tanpa `service:` tetap didukung sebagai key global.

### Token Session Gateway
- `ANY /wa/*` dengan `token` atau `Authorization: Bearer <token>`
- Path admin `'/wa/admin*'` diblokir agar tidak terekspos ke public.

## Security

- Rate limiter dan anti-spam berbasis IP aktif untuk API publik.
- Endpoint `/internal/*` dibypass dari limiter publik dan wajib `x-internal-api-key`.
- API key customer disimpan dalam bentuk hash SHA-256.
- Cron WIB (`Asia/Jakarta`) berjalan tiap menit untuk auto-set subscription `expired`.

## Menjalankan Service

1. Copy `.env.example` ke `.env`.
2. Isi variabel DB, `WA_SERVER_URL`, `WA_ADMIN_TOKEN`, `INTERNAL_API_KEYS`.
3. Jalankan `go run .` atau `docker compose up -d --build`.
