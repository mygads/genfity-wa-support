# Bulk Message API Documentation

API untuk mengirim pesan WhatsApp secara bulk (massal) dengan penjadwalan.

## Base URL
```
http://localhost:8070
```

## Authentication
Semua endpoint memerlukan header `token` yang berisi WhatsApp session token.

```
Headers:
token: your_whatsapp_session_token
```

---

## Endpoints

### 1. Create Bulk Text Message

Membuat kampanye pesan teks bulk.

**Endpoint:** `POST /bulk/create/text`

**Headers:**
```
Content-Type: application/json
token: your_whatsapp_session_token
```

**Request Body:**
```json
{
  "Phone": ["6281233784490", "6287327273773"],
  "Body": "Halo! Ini adalah pesan bulk dari sistem kami.",
  "SendSync": "now"
}
```

**SendSync Options:**
- `"now"` atau `"sekarang"` - Kirim sekarang
- `"2024-12-25 14:30:00"` - Kirim pada tanggal dan jam tertentu
- `"2024-12-25T14:30:00"` - Format ISO datetime
- `"25/12/2024 14:30"` - Format tanggal Indonesia

**Response:**
```json
{
  "code": 200,
  "success": true,
  "message": "Bulk text message created successfully",
  "data": {
    "bulk_id": 1,
    "total_recipients": 2,
    "status": "pending",
    "scheduled_at": null
  }
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8070/bulk/create/text \
  -H "Content-Type: application/json" \
  -H "token: your_session_token" \
  -d '{
    "Phone": ["6281233784490", "6287327273773"],
    "Body": "Halo! Ini pesan test bulk.",
    "SendSync": "now"
  }'
```

---

### 2. Create Bulk Image Message

Membuat kampanye pesan gambar bulk.

**Endpoint:** `POST /bulk/create/image`

**Headers:**
```
Content-Type: application/json
token: your_whatsapp_session_token
```

**Request Body (dengan URL):**
```json
{
  "Phone": ["6281233784490", "6287327273773"],
  "Image": "https://example.com/image.jpg",
  "Caption": "Ini adalah caption gambar",
  "SendSync": "2024-12-25 15:00:00"
}
```

**Request Body (dengan Base64):**
```json
{
  "Phone": ["6281233784490"],
  "Image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
  "Caption": "Gambar dari base64",
  "SendSync": "now"
}
```

**Response:**
```json
{
  "code": 200,
  "success": true,
  "message": "Bulk image message created successfully",
  "data": {
    "bulk_id": 2,
    "total_recipients": 1,
    "status": "scheduled",
    "scheduled_at": "2024-12-25T15:00:00Z"
  }
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8070/bulk/create/image \
  -H "Content-Type: application/json" \
  -H "token: your_session_token" \
  -d '{
    "Phone": ["6281233784490"],
    "Image": "https://picsum.photos/300/200",
    "Caption": "Test gambar bulk",
    "SendSync": "now"
  }'
```

---

### 3. Get Bulk Message List

Mendapatkan daftar semua bulk message dari user.

**Endpoint:** `GET /bulk/message`

**Headers:**
```
token: your_whatsapp_session_token
```

**Response:**
```json
{
  "code": 200,
  "success": true,
  "data": [
    {
      "id": 1,
      "session_id": "session123",
      "message_type": "text",
      "phone_numbers": ["6281233784490", "6287327273773"],
      "body": "Halo! Ini adalah pesan bulk.",
      "image": "",
      "caption": "",
      "send_sync": "now",
      "scheduled_at": null,
      "status": "completed",
      "total_recipients": 2,
      "sent_count": 2,
      "failed_count": 0,
      "processed_at": "2024-12-25T10:30:00Z",
      "completed_at": "2024-12-25T10:31:00Z",
      "error_message": "",
      "created_at": "2024-12-25T10:30:00Z",
      "updated_at": "2024-12-25T10:31:00Z"
    }
  ]
}
```

**Example cURL:**
```bash
curl -X GET http://localhost:8070/bulk/message \
  -H "token: your_session_token"
```

---

### 4. Get Bulk Message Detail

Mendapatkan detail bulk message beserta status pengiriman individual.

**Endpoint:** `GET /bulk/message/{id}`

**Headers:**
```
token: your_whatsapp_session_token
```

**Response:**
```json
{
  "code": 200,
  "success": true,
  "data": {
    "bulk_message": {
      "id": 1,
      "session_id": "session123",
      "message_type": "text",
      "phone_numbers": ["6281233784490", "6287327273773"],
      "body": "Halo! Ini adalah pesan bulk.",
      "send_sync": "now",
      "status": "completed",
      "total_recipients": 2,
      "sent_count": 2,
      "failed_count": 0,
      "created_at": "2024-12-25T10:30:00Z"
    },
    "items": [
      {
        "id": 1,
        "bulk_message_id": 1,
        "phone_number": "6281233784490",
        "status": "sent",
        "message_id": "msg_123456",
        "error_message": "",
        "sent_at": "2024-12-25T10:30:15Z",
        "created_at": "2024-12-25T10:30:00Z"
      },
      {
        "id": 2,
        "bulk_message_id": 1,
        "phone_number": "6287327273773",
        "status": "sent",
        "message_id": "msg_123457",
        "error_message": "",
        "sent_at": "2024-12-25T10:30:18Z",
        "created_at": "2024-12-25T10:30:00Z"
      }
    ]
  }
}
```

**Example cURL:**
```bash
curl -X GET http://localhost:8070/bulk/message/1 \
  -H "token: your_session_token"
```

---

### 5. Cron Job Endpoint

Endpoint untuk memproses pesan yang dijadwalkan (digunakan oleh cron job).

**Endpoint:** `GET /bulk/cron/process`

**Response:**
```json
{
  "code": 200,
  "success": true,
  "message": "Cron job completed",
  "data": {
    "processed_count": 3,
    "checked_at": "2024-12-25T10:30:00Z"
  }
}
```

**Cron Setup:**
Jalankan setiap menit untuk mengecek pesan terjadwal:
```bash
* * * * * curl http://localhost:8070/bulk/cron/process
```

---

## Status Codes

### Bulk Message Status
- `pending` - Menunggu diproses
- `scheduled` - Dijadwalkan untuk nanti
- `processing` - Sedang diproses
- `completed` - Selesai
- `failed` - Gagal

### Individual Message Status
- `pending` - Belum dikirim
- `sent` - Berhasil dikirim
- `failed` - Gagal dikirim

## Error Responses

**401 Unauthorized:**
```json
{
  "code": 401,
  "success": false,
  "message": "Token is required"
}
```

**400 Bad Request:**
```json
{
  "code": 400,
  "success": false,
  "message": "Invalid request format: validation error"
}
```

**500 Internal Server Error:**
```json
{
  "code": 500,
  "success": false,
  "message": "Failed to create bulk message: database error"
}
```

---

## Database Schema

### bulk_messages Table
```sql
CREATE TABLE bulk_messages (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR NOT NULL,
    message_type VARCHAR NOT NULL, -- 'text' or 'image'
    phone_numbers JSONB NOT NULL,
    body TEXT,
    image TEXT,
    caption TEXT,
    send_sync VARCHAR NOT NULL,
    scheduled_at TIMESTAMP,
    status VARCHAR DEFAULT 'pending',
    total_recipients INTEGER,
    sent_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    processed_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

### bulk_message_items Table
```sql
CREATE TABLE bulk_message_items (
    id SERIAL PRIMARY KEY,
    bulk_message_id INTEGER REFERENCES bulk_messages(id),
    phone_number VARCHAR NOT NULL,
    status VARCHAR DEFAULT 'pending',
    message_id VARCHAR,
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

---

## Integration dengan WhatsApp Server

Sistem ini akan terintegrasi dengan WhatsApp server yang ada di `{baseUrl}/bulk/create/text` dan `{baseUrl}/bulk/create/image` dengan:

1. **Header token** yang sama dari user session
2. **Payload** yang disesuaikan dengan format yang diharapkan server WhatsApp
3. **Response handling** untuk update status pengiriman
4. **Error handling** untuk kasus pengiriman gagal

---

## Workflow

1. **User membuat bulk message** via API endpoint
2. **Sistem menyimpan** ke database dengan status pending/scheduled
3. **Jika immediate (now)**: langsung diproses
4. **Jika scheduled**: menunggu cron job
5. **Cron job berjalan setiap menit** mengecek pesan terjadwal
6. **Sistem mengirim** ke WhatsApp server
7. **Update status** berdasarkan response dari WhatsApp server

---

## Testing

### Test Bulk Text Message
```bash
curl -X POST http://localhost:8070/bulk/create/text \
  -H "Content-Type: application/json" \
  -H "token: YOUR_TOKEN" \
  -d '{
    "Phone": ["6281234567890"],
    "Body": "Test pesan bulk",
    "SendSync": "now"
  }'
```

### Test Bulk Image Message  
```bash
curl -X POST http://localhost:8070/bulk/create/image \
  -H "Content-Type: application/json" \
  -H "token: YOUR_TOKEN" \
  -d '{
    "Phone": ["6281234567890"],
    "Image": "https://picsum.photos/200/300",
    "Caption": "Test gambar",
    "SendSync": "now"
  }'
```

### Test Scheduled Message
```bash
curl -X POST http://localhost:8070/bulk/create/text \
  -H "Content-Type: application/json" \
  -H "token: YOUR_TOKEN" \
  -d '{
    "Phone": ["6281234567890"],
    "Body": "Pesan terjadwal",
    "SendSync": "2024-12-25 15:30:00"
  }'
```
