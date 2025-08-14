# Testing API Endpoints

## 1. Test Webhook Message
```bash
curl -X POST http://localhost:8070/webhook/wa \
  -H "Content-Type: application/json" \
  -d '{
    "event": {
      "Info": {
        "ID": "TEST123456789",
        "Sender": "6281233784490@s.whatsapp.net",
        "Chat": "6289668176764@s.whatsapp.net",
        "IsFromMe": false,
        "IsGroup": false,
        "PushName": "Test User",
        "Type": "text",
        "Timestamp": "2025-08-14T15:00:00+07:00"
      },
      "Message": {
        "conversation": "Hello test message"
      }
    },
    "token": "genfitywa1",
    "type": "Message"
  }'
```

## 2. Test Session Sync (Main Feature)
```bash
curl -X GET http://localhost:8070/api/v1/sessions/sync
```

## 3. Get Sessions
```bash
curl -X GET http://localhost:8070/api/v1/sessions
```

## 4. Get Messages
```bash
curl -X GET "http://localhost:8070/api/v1/messages?user_token=genfitywa1&limit=5"
```

## 5. Get Message Statuses
```bash
curl -X GET "http://localhost:8070/api/v1/message-statuses?user_token=genfitywa1"
```

## 6. Get Chat Presences (Typing Status)
```bash
curl -X GET "http://localhost:8070/api/v1/chat-presences?user_token=genfitywa1"
```

## 7. Get QR Code for Session
```bash
curl -X GET http://localhost:8070/api/v1/sessions/genfitywa1/qr
```

## 8. Test ReadReceipt Webhook
```bash
curl -X POST http://localhost:8070/webhook/wa \
  -H "Content-Type: application/json" \
  -d '{
    "event": {
      "Chat": "6289668176764@s.whatsapp.net",
      "Sender": "6281233784490@s.whatsapp.net",
      "MessageIDs": ["TEST123456789"],
      "Timestamp": "2025-08-14T15:01:00+07:00",
      "Type": "read"
    },
    "state": "Read",
    "token": "genfitywa1",
    "type": "ReadReceipt"
  }'
```

## 9. Test ChatPresence Webhook (Typing)
```bash
curl -X POST http://localhost:8070/webhook/wa \
  -H "Content-Type: application/json" \
  -d '{
    "event": {
      "Chat": "6289668176764@s.whatsapp.net",
      "Sender": "6281233784490@s.whatsapp.net",
      "State": "composing"
    },
    "token": "genfitywa1",
    "type": "ChatPresence"
  }'
```

## 10. Health Check
```bash
curl -X GET http://localhost:8070/api/v1/health
```

## Environment Setup untuk .env
```
# WhatsApp Server Configuration (REQUIRED for sync)
WA_SERVER_URL=http://your-wa-server:8080
WA_ADMIN_TOKEN=your_admin_token_here

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=genfity
DB_PASSWORD=genfitywa
DB_NAME=whatsapp_events
DB_SSLMODE=disable

# Server configuration
PORT=8070
```

## Cron Job Setup
Untuk monitoring otomatis, tambahkan ke crontab:
```bash
# Check session status every 30 seconds
* * * * * curl -s http://localhost:8070/api/v1/sessions/sync > /dev/null
* * * * * sleep 30; curl -s http://localhost:8070/api/v1/sessions/sync > /dev/null
```

Atau untuk interval yang lebih lama (setiap 5 menit):
```bash
*/5 * * * * curl -s http://localhost:8070/api/v1/sessions/sync > /dev/null
```

## Response Format Examples

### Session Sync Response:
```json
{
  "status": "success",
  "message": "Session status synced successfully",
  "stats": {
    "total_sessions": 2,
    "created_sessions": 0,
    "updated_sessions": 2,
    "last_sync_at": "2025-08-14T15:00:00Z"
  },
  "sessions": [
    {
      "connected": true,
      "loggedIn": true,
      "token": "genfitywa1",
      "jid": "6289668176764:77@s.whatsapp.net"
    }
  ]
}
```

### Messages Response:
```json
{
  "messages": [
    {
      "id": 1,
      "message_id": "TEST123456789",
      "from": "6281233784490@s.whatsapp.net",
      "to": "6289668176764@s.whatsapp.net",
      "body": "Hello test message",
      "message_type": "text",
      "status": "read",
      "user_token": "genfitywa1"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```
