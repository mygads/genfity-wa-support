# Genfity Chat AI - WhatsApp Event API & Gateway

A comprehensive WhatsApp API gateway and event processing system with subscription validation, image processing, contact management, and bulk messaging capabilities.

## Features

### Core Gateway Features
- ✅ **API Gateway** - Proxy requests to WhatsApp server with validation
- ✅ **Subscription Validation** - Check user subscription status  
- ✅ **Session Management** - Validate and track WhatsApp sessions
- ✅ **Message Tracking** - Track message quotas and limits
- ✅ **Image Processing** - URL to base64 conversion for images
- ✅ **Admin Bypass** - Admin endpoints bypass all validation

### Contact Management
- ✅ **Contact Sync** - Sync contacts from WhatsApp server
- ✅ **Contact List** - Get simplified contact list
- ✅ **Database Storage** - Store contacts with session linkage

### Bulk Messaging System
- ✅ **Bulk Text Messages** - Send text messages to multiple recipients
- ✅ **Bulk Image Messages** - Send images (URL/base64) to multiple recipients  
- ✅ **Message Scheduling** - Schedule messages for future delivery
- ✅ **Status Tracking** - Track individual message delivery status
- ✅ **Cron Job Support** - Process scheduled messages automatically

### Event Processing
- ✅ **Webhook Handling** - Receive and process WhatsApp events
- ✅ **Event Storage** - Store all events with proper categorization
- ✅ **Real-time Processing** - Process events in real-time
- ✅ **Event Types** - Support for messages, receipts, presence, etc.

## API Documentation

### Gateway Endpoints
- **All WA Routes**: `/wa/*` - Proxied to WhatsApp server with validation
- **Health Check**: `GET /health` - Server status

### Contact Management
- **Sync Contacts**: `POST /bulk/contact/sync` - Sync from WhatsApp server
- **List Contacts**: `GET /bulk/contact` - Get simplified contact list

📖 **[Complete Contact API Documentation](./CONTACT_SYNC_API.md)**

### Bulk Messaging
- **Create Text Bulk**: `POST /bulk/create/text` - Create bulk text campaign
- **Create Image Bulk**: `POST /bulk/create/image` - Create bulk image campaign
- **List Campaigns**: `GET /bulk/message` - Get all bulk campaigns
- **Campaign Detail**: `GET /bulk/message/{id}` - Get campaign details
- **Cron Process**: `GET /bulk/cron/process` - Process scheduled messages

📖 **[Complete Bulk Message API Documentation](./BULK_MESSAGE_API.md)**

### Webhook Events
- **Verify Webhook**: `GET /webhook/wa` - WhatsApp webhook verification
- **Handle Events**: `POST /webhook/wa` - Process incoming WhatsApp events

📖 **[Webhook Documentation](./WEBHOOK_DOCUMENTATION.md)**

## Quick Start

### 1. Environment Setup
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=chat-ai_db
DB_SSLMODE=disable

# Transactional Database
TRANSACTIONAL_DB_HOST=localhost
TRANSACTIONAL_DB_PORT=5432
TRANSACTIONAL_DB_USER=postgres
TRANSACTIONAL_DB_PASSWORD=password
TRANSACTIONAL_DB_NAME=transactional_db
TRANSACTIONAL_DB_SSLMODE=disable

# WhatsApp Server
WHATSAPP_SERVER_URL=https://wa.genfity.com

# Gateway Configuration
GATEWAY_MODE=production
PORT=8070
```

### 2. Run the Server
```bash
go run main.go
```

### 3. Test Contact Sync
```bash
curl -X POST http://localhost:8070/bulk/contact/sync \
  -H "token: your_session_token"
```

### 4. Create Bulk Message
```bash
curl -X POST http://localhost:8070/bulk/create/text \
  -H "Content-Type: application/json" \
  -H "token: your_session_token" \
  -d '{
    "Phone": ["6281234567890"],
    "Body": "Hello from bulk messaging!",
    "SendSync": "now"
  }'
```

## Database Schema

### Primary Database (chat-ai_db)
- `gen_event_webhooks` - All webhook events
- `whats_app_messages` - Processed messages
- `whats_app_sessions` - Local session tracking
- `user_settings` - User preferences
- `chat_rooms` - Chat conversations
- `chat_messages` - Individual messages

### Transactional Database
- `WhatsAppSession` - Session management
- `WhatsappApiPackage` - Package configuration  
- `ServicesWhatsappCustomers` - Subscription data
- `WhatsAppMessageStats` - Message statistics
- `whats_app_sync_contacts` - Synced contacts
- `bulk_messages` - Bulk message campaigns
- `bulk_message_items` - Individual message items

## Gateway Routing

### Admin Routes (No Validation)
```
/wa/admin/*     → {baseUrl}/admin/*
```

### Validated Routes
```
/wa/session/*   → {baseUrl}/session/*   (+ session limits)
/wa/webhook/*   → {baseUrl}/webhook/*   (+ subscription)
/wa/chat/*      → {baseUrl}/chat/*      (+ message tracking)
/wa/user/*      → {baseUrl}/user/*      (+ subscription)
/wa/group/*     → {baseUrl}/group/*     (+ subscription)
/wa/newsletter/* → {baseUrl}/newsletter/* (+ subscription)
```

### Special Processing
- **Image Endpoints**: Automatic URL → base64 conversion
- **Token Validation**: All routes validate session token
- **Subscription Check**: Non-admin routes check subscription status
- **Message Tracking**: Chat routes track message quotas

## Bulk Messaging Workflow

1. **Create Campaign** - User creates text/image bulk campaign
2. **Schedule Processing** - Immediate or scheduled for later
3. **Contact Validation** - Validate phone numbers
4. **Queue Processing** - Process messages in background
5. **Status Updates** - Track individual delivery status
6. **Cron Jobs** - Process scheduled campaigns every minute

## Contact Sync Workflow

1. **Trigger Sync** - User calls sync endpoint
2. **External Request** - Fetch from WhatsApp server
3. **Data Processing** - Clean and validate contact data
4. **Database Storage** - Store with session linkage
5. **Response** - Return synced contact data

## Error Handling

- **401 Unauthorized** - Invalid or missing token
- **403 Forbidden** - Subscription expired or limit exceeded
- **400 Bad Request** - Invalid request format
- **500 Internal Error** - Server or database errors

## Logging

- **GORM Silent Mode** - Clean logs without query noise
- **Error Tracking** - Comprehensive error logging
- **Performance Metrics** - Request timing and success rates

## Contributing

1. Follow Go best practices
2. Add tests for new features
3. Update documentation
4. Ensure proper error handling

## License

This project is licensed under the MIT License.
