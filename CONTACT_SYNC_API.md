# Contact Sync API Documentation

API untuk sinkronisasi dan pengelolaan kontak WhatsApp.

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

### 1. Bulk Contact Sync

Melakukan sinkronisasi kontak dari server WhatsApp eksternal dan menyimpannya ke database.

**Endpoint:** `POST /bulk/contact/sync`

**Headers:**
```
token: your_whatsapp_session_token
```

**Response:**
```json
{
  "code": 200,
  "success": true,
  "message": "Successfully synced 390 contacts",
  "data": {
    "136202103562334@lid": {
      "BusinessName": "",
      "FirstName": "",
      "Found": true,
      "FullName": "",
      "PushName": "Ezra Balqis Alka Ceria"
    },
    "6285103879393@s.whatsapp.net": {
      "BusinessName": "",
      "FirstName": "Dr. Pikir Wisnu Wijayanto",
      "Found": true,
      "FullName": "Dr. Pikir Wisnu Wijayanto Dosen Inggris",
      "PushName": ""
    }
  },
  "stored": 390
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8070/bulk/contact/sync \
  -H "token: your_session_token"
```

**What happens:**
1. API mengambil kontak dari `{baseUrl}/user/contacts` 
2. Menyimpan/update kontak ke database transactional
3. Mengembalikan data kontak yang berhasil disinkronisasi

---

### 2. Get Bulk Contact List

Mendapatkan daftar kontak yang telah disinkronisasi dalam format yang disederhanakan.

**Endpoint:** `GET /bulk/contact`

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
      "telp": "6285215538030",
      "fullname": "Fathi Rizki"
    },
    {
      "telp": "6285103879393", 
      "fullname": "Dr. Pikir Wisnu Wijayanto Dosen Inggris"
    },
    {
      "telp": "6285117222016",
      "fullname": "Abdul Kim"
    }
  ]
}
```

**Example cURL:**
```bash
curl -X GET http://localhost:8070/bulk/contact \
  -H "token: your_session_token"
```

**Data Processing:**
- Mengekstrak nomor telepon dari WhatsApp JID
- Memilih nama terbaik dengan prioritas: FullName > PushName > FirstName > BusinessName
- Hanya menampilkan kontak yang memiliki nama dan nomor telepon
- Mengurutkan data kontak yang ditemukan (`Found: true`)

---

## Integration Flow

### Contact Sync Workflow
1. **User memanggil** `/bulk/contact/sync`
2. **Sistem melakukan request** ke WhatsApp server eksternal: `{baseUrl}/user/contacts`
3. **Server WhatsApp mengembalikan** data kontak dalam format JSON
4. **Sistem memproses** data dan menyimpan ke database `whats_app_sync_contacts`
5. **Response** dikembalikan dengan data asli + informasi penyimpanan

### Contact List Workflow  
1. **User memanggil** `/bulk/contact`
2. **Sistem mengambil** kontak dari database berdasarkan session ID
3. **Data difilter** dan diformat ulang
4. **Response** dikembalikan dalam format sederhana

---

## Database Schema

### whats_app_sync_contacts Table
```sql
CREATE TABLE whats_app_sync_contacts (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR NOT NULL,
    contact_jid VARCHAR NOT NULL, -- WhatsApp ID seperti "628xxx@s.whatsapp.net"
    business_name VARCHAR,
    first_name VARCHAR,
    full_name VARCHAR,
    push_name VARCHAR,
    found BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    INDEX(session_id),
    INDEX(contact_jid),
    UNIQUE(session_id, contact_jid)
);
```

---

## Data Format

### WhatsApp JID Format
- **Phone numbers:** `6285215538030@s.whatsapp.net`
- **Local numbers:** `136202103562334@lid`

### Contact Name Priority
1. **FullName** - Nama lengkap jika tersedia
2. **PushName** - Nama yang ditampilkan di WhatsApp
3. **FirstName** - Nama depan
4. **BusinessName** - Nama bisnis

---

## Error Handling

### 401 Unauthorized
```json
{
  "code": 401,
  "success": false,
  "message": "Invalid token"
}
```

### 500 Internal Server Error
```json
{
  "code": 500,
  "success": false,
  "message": "Failed to connect to WhatsApp server"
}
```

### WhatsApp Server Error
```json
{
  "code": 400,
  "success": false,  
  "message": "Failed to sync contacts from WhatsApp server"
}
```

---

## Usage Examples

### Complete Contact Sync Flow
```bash
# 1. Sync contacts from WhatsApp server
curl -X POST http://localhost:8070/bulk/contact/sync \
  -H "token: cmf16gx9f0033jt28qx6slviy"

# 2. Get simplified contact list  
curl -X GET http://localhost:8070/bulk/contact \
  -H "token: cmf16gx9f0033jt28qx6slviy"

# 3. Use contact data for bulk messaging
curl -X POST http://localhost:8070/bulk/create/text \
  -H "Content-Type: application/json" \
  -H "token: cmf16gx9f0033jt28qx6slviy" \
  -d '{
    "Phone": ["6285215538030", "6285103879393"],
    "Body": "Halo! Pesan dari sistem kami.",
    "SendSync": "now"
  }'
```

---

## Best Practices

1. **Sync contacts regularly** untuk memastikan data terbaru
2. **Check response status** sebelum menggunakan data
3. **Handle pagination** jika diperlukan untuk kontak dalam jumlah besar
4. **Implement retry logic** untuk kasus network error
5. **Validate phone numbers** sebelum mengirim bulk message

---

## Integration Notes

- **Database:** Menggunakan database transactional yang sama dengan session management
- **Authentication:** Token harus valid di table `WhatsAppSession`
- **External API:** Bergantung pada `{baseUrl}/user/contacts` endpoint dari WhatsApp server
- **Performance:** Sync bisa memakan waktu tergantung jumlah kontak
- **Storage:** Data disimpan dengan soft delete untuk audit trail
