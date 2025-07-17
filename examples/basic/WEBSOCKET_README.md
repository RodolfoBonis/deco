# Deco WebSocket Demo 🚀

This example demonstrates comprehensive WebSocket functionality using the Deco framework, including real-time chat, notifications, presence tracking, and live updates.

## 🎯 Features Demonstrated

### WebSocket Endpoints
- **Main WebSocket**: `/ws` - General purpose WebSocket connection
- **Chat Rooms**: `/ws/chat/:room` - Room-based chat functionality
- **Notifications**: `/ws/notifications` - Push notification channel
- **Live Updates**: `/ws/live` - Real-time data updates

### HTTP API Endpoints
- **Chat Management**: Send messages via HTTP API
- **Notification System**: Send push notifications
- **Presence Tracking**: Update and query user status
- **Statistics**: Monitor WebSocket connections and usage

### Schema Types
- `ChatMessage` - Real-time chat messages
- `NotificationMessage` - Push notifications
- `PresenceInfo` - User presence and status
- `LiveUpdateData` - Real-time data updates

## 🚀 Getting Started

### 1. Start the Server

```bash
# From the examples/basic directory
../../deco generate
./basic
```

The server will start on `http://localhost:8080`

### 2. Test WebSocket Functionality

#### Option A: Web Interface
Open your browser and navigate to:
```
http://localhost:8080/demo/websocket
```

#### Option B: Standalone Test Client
Open the `websocket_test.html` file in your browser for a comprehensive testing interface.

#### Option C: Command Line Testing
Use tools like `wscat` or `websocat`:

```bash
# Install wscat
npm install -g wscat

# Connect to main WebSocket
wscat -c ws://localhost:8080/ws?user_id=test_user

# Connect to chat room
wscat -c "ws://localhost:8080/ws/chat/general?user_id=test_user&username=TestUser"

# Connect to notifications
wscat -c ws://localhost:8080/ws/notifications?user_id=test_user
```

## 📡 WebSocket Message Format

All WebSocket messages follow this JSON structure:

```json
{
  "type": "message_type",
  "data": {
    // Message-specific data
  },
  "sender": "connection_id",
  "target": "target_connection_id",  // Optional
  "group": "group_name",             // Optional
  "timestamp": "2023-12-07T10:30:00Z",
  "metadata": {}                     // Optional
}
```

### Message Types

#### Chat Messages
```json
{
  "type": "chat",
  "data": {
    "id": "msg_1701944200000",
    "user_id": "user123",
    "username": "Alice",
    "room": "general",
    "message": "Hello everyone!",
    "timestamp": "2023-12-07T10:30:00Z"
  }
}
```

#### Notifications
```json
{
  "type": "notification",
  "data": {
    "id": "notif_1701944200000",
    "title": "New Message",
    "body": "You have a new message from Bob",
    "type": "info",
    "priority": "normal",
    "user_id": "user123"
  }
}
```

#### Presence Updates
```json
{
  "type": "presence",
  "data": {
    "user_id": "user123",
    "username": "Alice",
    "status": "online",
    "last_seen": "2023-12-07T10:30:00Z",
    "room": "general"
  }
}
```

#### Live Updates
```json
{
  "type": "live_update",
  "data": {
    "type": "user_count",
    "resource": "users",
    "data": {"count": 42},
    "timestamp": "2023-12-07T10:30:00Z"
  }
}
```

## 🔧 HTTP API Endpoints

### Chat API

#### Send Chat Message
```bash
POST /api/chat/send
Content-Type: application/json

{
  "user_id": "user123",
  "username": "Alice",
  "room": "general",
  "message": "Hello from HTTP API!"
}
```

#### Get Chat Rooms
```bash
GET /api/chat/rooms

Response:
{
  "rooms": [
    {
      "name": "general",
      "user_count": 5,
      "users": ["user1", "user2", "user3"]
    }
  ],
  "total": 1
}
```

### Notification API

#### Send Notification
```bash
POST /api/notifications/send
Content-Type: application/json

{
  "title": "System Alert",
  "body": "Server maintenance in 10 minutes",
  "type": "warning",
  "priority": "high",
  "user_id": "user123"  // Optional: specific user, omit for broadcast
}
```

### Presence API

#### Update Presence
```bash
POST /api/presence/update
Content-Type: application/json

{
  "user_id": "user123",
  "username": "Alice",
  "status": "online",
  "room": "general"
}
```

#### Get Online Users
```bash
GET /api/presence/online?room=general

Response:
{
  "users": [
    {
      "user_id": "user123",
      "username": "Alice",
      "status": "online",
      "last_seen": "2023-12-07T10:30:00Z",
      "room": "general"
    }
  ],
  "count": 1
}
```

### Live Updates API

#### Send Live Update
```bash
POST /api/live/update
Content-Type: application/json

{
  "type": "user_count",
  "resource": "users",
  "data": {"count": 42}
}
```

### Monitoring API

#### Get WebSocket Statistics
```bash
GET /api/websocket/stats

Response:
{
  "total_connections": 15,
  "active_rooms": 3,
  "online_users": 12,
  "timestamp": "2023-12-07T10:30:00Z"
}
```

## 🎮 Testing Scenarios

### Scenario 1: Multi-User Chat
1. Open multiple browser tabs to `/demo/websocket`
2. Connect with different usernames to the same room
3. Send messages and see real-time updates

### Scenario 2: Notification Broadcasting
1. Connect multiple clients to `/ws/notifications`
2. Use the HTTP API to send notifications
3. Observe real-time delivery

### Scenario 3: Presence Tracking
1. Connect users to different rooms
2. Update presence status via API
3. Monitor online user lists

### Scenario 4: Live Data Updates
1. Connect clients to `/ws/live`
2. Send live updates via API
3. See real-time data changes

## 🛠️ Configuration

WebSocket settings in `.deco.yaml`:

```yaml
websocket:
  enabled: true           # Enable WebSocket support
  read_buffer: 1024      # Read buffer size
  write_buffer: 1024     # Write buffer size
  check_origin: false    # CORS origin checking
  compression: true      # Enable compression
  ping_interval: 54s     # Ping interval
  pong_timeout: 60s      # Pong timeout
```

## 📊 Decorators Used

### @WebSocket()
Marks a handler as a WebSocket endpoint:
```go
// @WebSocket()
func HandleWebSocketConnection(c *gin.Context) {
    // WebSocket logic handled by framework
}
```

### @WebSocketStats()
Enables WebSocket statistics collection:
```go
// @WebSocketStats()
func GetWebSocketStats(c *gin.Context) {
    // Statistics automatically populated
}
```

## 🔍 Troubleshooting

### Connection Issues
- Verify server is running on correct port
- Check WebSocket is enabled in configuration
- Ensure firewall allows WebSocket connections

### Message Not Received
- Verify correct message format
- Check user is connected to correct endpoint
- Verify room/group membership for targeted messages

### Performance Issues
- Monitor connection count in statistics
- Check buffer sizes in configuration
- Consider enabling compression for large messages

## 🚀 Advanced Usage

### Custom Message Handlers
The framework automatically routes messages based on type. You can extend functionality by:

1. Adding new message types
2. Implementing custom business logic
3. Integrating with external services

### Scaling Considerations
For production deployments:

1. Use Redis for session storage
2. Implement horizontal scaling
3. Add load balancing for WebSocket connections
4. Monitor connection metrics

## 📝 Code Examples

### JavaScript Client
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/chat/room1?user_id=user123&username=Alice');

ws.onopen = function() {
    console.log('Connected to WebSocket');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};

// Send a chat message
const message = {
    type: 'chat',
    data: {
        user_id: 'user123',
        username: 'Alice',
        room: 'room1',
        message: 'Hello World!',
        timestamp: new Date().toISOString()
    }
};
ws.send(JSON.stringify(message));
```

### Go Client
```go
package main

import (
    "log"
    "github.com/gorilla/websocket"
)

func main() {
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
    if err != nil {
        log.Fatal("Dial error:", err)
    }
    defer conn.Close()

    // Read messages
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }
        log.Printf("Received: %s", message)
    }
}
```

---

This comprehensive WebSocket demo showcases the power and flexibility of the Deco framework for building real-time applications! 🎉 