package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/RodolfoBonis/deco/pkg/decorators"
	"github.com/gin-gonic/gin"
)

// WebSocket message types
const (
	MsgTypeChat         = "chat"
	MsgTypeNotification = "notification"
	MsgTypePresence     = "presence"
	MsgTypeJoinRoom     = "join_room"
	MsgTypeLeaveRoom    = "leave_room"
	MsgTypeUserList     = "user_list"
	MsgTypeLiveUpdate   = "live_update"
	MsgTypeHeartbeat    = "heartbeat"
)

// ChatMessage represents a chat message
// @Schema
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Room      string    `json:"room"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// NotificationMessage represents a push notification
// @Schema
type NotificationMessage struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Type     string                 `json:"type"`
	Priority string                 `json:"priority"` // low, normal, high
	Data     map[string]interface{} `json:"data,omitempty"`
	UserID   string                 `json:"user_id,omitempty"`
}

// PresenceInfo represents user presence information
// @Schema
type PresenceInfo struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Status   string    `json:"status"` // online, away, busy, offline
	LastSeen time.Time `json:"last_seen"`
	Room     string    `json:"room,omitempty"`
	Avatar   string    `json:"avatar,omitempty"`
}

// LiveUpdateData represents real-time data updates
// @Schema
type LiveUpdateData struct {
	Type      string      `json:"type"`     // user_count, new_order, status_change
	Resource  string      `json:"resource"` // users, orders, systems
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// In-memory storage for demo (in production, use Redis or database)
var (
	activeUsers = make(map[string]*PresenceInfo)
	chatRooms   = make(map[string][]string) // room -> user_ids
	onlineCount = 0
)

// =============================================================================
// WebSocket Connection Handlers
// =============================================================================

// HandleWebSocketConnection handles main WebSocket connections
// @Route("GET", "/ws")
// @Summary("WebSocket connection endpoint")
// @Description("Establishes WebSocket connection for real-time communication")
// @Tag("websocket")
// @WebSocket()
// @Response(101, description="Switching protocols to WebSocket")
func HandleWebSocketConnection(c *gin.Context) {
	// This will be handled by the @WebSocket decorator
	// The actual WebSocket logic is managed by the framework
	c.JSON(http.StatusUpgradeRequired, gin.H{
		"message":  "Use WebSocket client to connect",
		"endpoint": "/ws",
	})
}

// HandleChatWebSocket handles chat-specific WebSocket connections
// @Route("GET", "/ws/chat/:room")
// @Summary("Chat room WebSocket connection")
// @Description("WebSocket connection for specific chat room")
// @Tag("websocket")
// @Tag("chat")
// @WebSocket()
// @Response(101, description="WebSocket connection established")
func HandleChatWebSocket(c *gin.Context) {
	room := c.Param("room")
	userID := c.Query("user_id")
	username := c.Query("username")

	if userID == "" || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id and username are required",
		})
		return
	}

	// Set context for WebSocket handler
	c.Set("room", room)
	c.Set("user_id", userID)
	c.Set("username", username)

	// WebSocket will be handled by decorator
	c.JSON(http.StatusUpgradeRequired, gin.H{
		"message": "Connecting to chat room",
		"room":    room,
	})
}

// HandleNotificationWebSocket handles notification WebSocket connections
// @Route("GET", "/ws/notifications")
// @Summary("Notification WebSocket connection")
// @Description("WebSocket connection for receiving real-time notifications")
// @Tag("websocket")
// @Tag("notifications")
// @WebSocket()
// @Response(101, description="WebSocket connection established")
func HandleNotificationWebSocket(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	c.Set("user_id", userID)
	c.Set("notification_channel", true)

	c.JSON(http.StatusUpgradeRequired, gin.H{
		"message": "Connecting to notification service",
		"user_id": userID,
	})
}

// HandleLiveUpdatesWebSocket handles live data updates
// @Route("GET", "/ws/live")
// @Summary("Live updates WebSocket connection")
// @Description("WebSocket connection for real-time data updates")
// @Tag("websocket")
// @Tag("live-updates")
// @WebSocket()
// @Response(101, description="WebSocket connection established")
func HandleLiveUpdatesWebSocket(c *gin.Context) {
	c.Set("live_updates", true)

	c.JSON(http.StatusUpgradeRequired, gin.H{
		"message": "Connecting to live updates",
	})
}

// =============================================================================
// HTTP API Endpoints for WebSocket Management
// =============================================================================

// SendChatMessage sends a chat message via HTTP (for testing)
// @Route("POST", "/api/chat/send")
// @Summary("Send chat message")
// @Description("Send a message to a chat room via HTTP API")
// @Tag("chat")
// @RequestBody(type="ChatMessage", description="Chat message data")
// @Response(200, description="Message sent successfully")
// @Response(400, description="Invalid message data")
func SendChatMessage(c *gin.Context) {
	var message ChatMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	message.Timestamp = time.Now()
	message.Type = MsgTypeChat

	// Broadcast to room via WebSocket
	broadcastToRoom(message.Room, &decorators.WebSocketMessage{
		Type:      MsgTypeChat,
		Data:      message,
		Group:     message.Room,
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Message sent successfully",
		"id":      message.ID,
	})
}

// SendNotification sends a notification via HTTP
// @Route("POST", "/api/notifications/send")
// @Summary("Send notification")
// @Description("Send a push notification to user(s)")
// @Tag("notifications")
// @RequestBody(type="NotificationMessage", description="Notification data")
// @Response(200, description="Notification sent successfully")
// @Response(400, description="Invalid notification data")
func SendNotification(c *gin.Context) {
	var notification NotificationMessage
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification.ID = fmt.Sprintf("notif_%d", time.Now().UnixNano())

	// Send via WebSocket
	if notification.UserID != "" {
		// Send to specific user
		sendToUser(notification.UserID, &decorators.WebSocketMessage{
			Type:      MsgTypeNotification,
			Data:      notification,
			Target:    notification.UserID,
			Timestamp: time.Now(),
		})
	} else {
		// Broadcast to all notification subscribers
		broadcastToGroup("notifications", &decorators.WebSocketMessage{
			Type:      MsgTypeNotification,
			Data:      notification,
			Timestamp: time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification sent successfully",
		"id":      notification.ID,
	})
}

// UpdatePresence updates user presence status
// @Route("POST", "/api/presence/update")
// @Summary("Update user presence")
// @Description("Update user online presence status")
// @Tag("presence")
// @Response(200, description="Presence updated successfully")
func UpdatePresence(c *gin.Context) {
	var presence PresenceInfo
	if err := c.ShouldBindJSON(&presence); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	presence.LastSeen = time.Now()
	activeUsers[presence.UserID] = &presence

	// Broadcast presence update
	broadcastToAll(&decorators.WebSocketMessage{
		Type:      MsgTypePresence,
		Data:      presence,
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Presence updated successfully",
	})
}

// GetOnlineUsers returns list of online users
// @Route("GET", "/api/presence/online")
// @Summary("Get online users")
// @Description("Get list of currently online users")
// @Tag("presence")
// @Response(200, description="Online users list")
func GetOnlineUsers(c *gin.Context) {
	room := c.Query("room")

	var users []PresenceInfo
	for _, user := range activeUsers {
		if room == "" || user.Room == room {
			users = append(users, *user)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
		"room":  room,
	})
}

// GetChatRooms returns available chat rooms
// @Route("GET", "/api/chat/rooms")
// @Summary("Get chat rooms")
// @Description("Get list of available chat rooms")
// @Tag("chat")
// @Response(200, description="Chat rooms list")
func GetChatRooms(c *gin.Context) {
	rooms := make([]map[string]interface{}, 0)

	for room, userIDs := range chatRooms {
		rooms = append(rooms, map[string]interface{}{
			"name":       room,
			"user_count": len(userIDs),
			"users":      userIDs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
		"total": len(rooms),
	})
}

// SendLiveUpdate sends live data update
// @Route("POST", "/api/live/update")
// @Summary("Send live update")
// @Description("Send real-time data update to connected clients")
// @Tag("live-updates")
// @RequestBody(type="LiveUpdateData", description="Live update data")
// @Response(200, description="Update sent successfully")
func SendLiveUpdate(c *gin.Context) {
	var update LiveUpdateData
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update.Timestamp = time.Now()

	// Broadcast live update
	broadcastToGroup("live_updates", &decorators.WebSocketMessage{
		Type:      MsgTypeLiveUpdate,
		Data:      update,
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Live update sent successfully",
	})
}

// =============================================================================
// WebSocket Statistics and Monitoring
// =============================================================================

// GetWebSocketStats returns WebSocket connection statistics
// @Route("GET", "/api/websocket/stats")
// @Summary("WebSocket statistics")
// @Description("Get current WebSocket connection statistics")
// @Tag("websocket")
// @Tag("monitoring")
// @WebSocketStats()
// @Response(200, description="WebSocket statistics")
func GetWebSocketStats(c *gin.Context) {
	// O middleware @WebSocketStats() já responde automaticamente
}

// =============================================================================
// Demo/Test Endpoints
// =============================================================================

// WebSocketDemo serves WebSocket demo page
// @Route("GET", "/demo/websocket")
// @Summary("WebSocket demo page")
// @Description("Serves HTML page for testing WebSocket functionality")
// @Tag("demo")
// @Response(200, description="Demo page served")
func WebSocketDemo(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Deco WebSocket Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .chat-box { border: 1px solid #ccc; height: 300px; overflow-y: scroll; padding: 10px; margin: 10px 0; }
        .input-group { margin: 10px 0; }
        input, button, select { padding: 8px; margin: 5px; }
        .status { padding: 10px; background: #f0f0f0; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Deco WebSocket Demo</h1>
        
        <div class="status" id="status">Disconnected</div>
        
        <div class="input-group">
            <input type="text" id="username" placeholder="Username" value="user_` + strconv.Itoa(int(time.Now().Unix()%1000)) + `">
            <input type="text" id="room" placeholder="Room" value="general">
            <button onclick="connect()">Connect</button>
            <button onclick="disconnect()">Disconnect</button>
        </div>
        
        <div class="input-group">
            <input type="text" id="message" placeholder="Type message..." style="width: 60%;">
            <button onclick="sendMessage()">Send</button>
        </div>
        
        <div class="chat-box" id="messages"></div>
        
        <div class="input-group">
            <h3>Actions:</h3>
            <button onclick="sendNotification()">Send Notification</button>
            <button onclick="updatePresence()">Update Presence</button>
            <button onclick="sendLiveUpdate()">Send Live Update</button>
        </div>
    </div>

    <script>
        let ws = null;
        let username = '';
        let room = '';

        function connect() {
            username = document.getElementById('username').value;
            room = document.getElementById('room').value;
            
            if (!username || !room) {
                alert('Please enter username and room');
                return;
            }
            
            const wsUrl = 'ws://localhost:8080/ws/chat/' + room + '?user_id=' + username + '&username=' + username;
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                document.getElementById('status').innerHTML = 'Connected to room: ' + room;
                document.getElementById('status').style.background = '#d4edda';
            };
            
            ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                addMessage(data);
            };
            
            ws.onclose = function() {
                document.getElementById('status').innerHTML = 'Disconnected';
                document.getElementById('status').style.background = '#f8d7da';
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
        }
        
        function disconnect() {
            if (ws) {
                ws.close();
            }
        }
        
        function sendMessage() {
            const messageInput = document.getElementById('message');
            const message = messageInput.value;
            
            if (!message || !ws) return;
            
            const data = {
                type: 'chat',
                data: {
                    user_id: username,
                    username: username,
                    room: room,
                    message: message,
                    timestamp: new Date().toISOString()
                }
            };
            
            ws.send(JSON.stringify(data));
            messageInput.value = '';
        }
        
        function addMessage(data) {
            const messages = document.getElementById('messages');
            const messageDiv = document.createElement('div');
            messageDiv.innerHTML = '<strong>' + data.type + '</strong>: ' + JSON.stringify(data.data, null, 2);
            messages.appendChild(messageDiv);
            messages.scrollTop = messages.scrollHeight;
        }
        
        function sendNotification() {
            fetch('/api/notifications/send', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    title: 'Test Notification',
                    body: 'This is a test notification from ' + username,
                    type: 'info',
                    priority: 'normal'
                })
            });
        }
        
        function updatePresence() {
            fetch('/api/presence/update', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    user_id: username,
                    username: username,
                    status: 'online',
                    room: room
                })
            });
        }
        
        function sendLiveUpdate() {
            fetch('/api/live/update', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    type: 'user_count',
                    resource: 'users',
                    data: {count: Math.floor(Math.random() * 100)}
                })
            });
        }
        
        // Enter key to send message
        document.getElementById('message').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// =============================================================================
// Helper Functions
// =============================================================================

func broadcastToRoom(room string, message *decorators.WebSocketMessage) {
	// This would interact with the WebSocket hub
	// Implementation depends on how the framework exposes the hub
	log.Printf("Broadcasting to room %s: %+v", room, message)
}

func sendToUser(userID string, message *decorators.WebSocketMessage) {
	// Send message to specific user
	log.Printf("Sending to user %s: %+v", userID, message)
}

func broadcastToGroup(group string, message *decorators.WebSocketMessage) {
	// Broadcast to all users in group
	log.Printf("Broadcasting to group %s: %+v", group, message)
}

func broadcastToAll(message *decorators.WebSocketMessage) {
	// Broadcast to all connected users
	log.Printf("Broadcasting to all: %+v", message)
}

// =============================================================================
// WebSocket Message Handlers (New Pattern)
// =============================================================================
// HandleChatMessage handles chat-type WebSocket messages
// @WebSocket("chat")
func HandleChatMessage(conn *decorators.WebSocketConnection, message *decorators.WebSocketMessage) error {
	log.Printf("📱 Chat message received and processed: %+v", message.Data)

	// Extract chat data from message
	if data, ok := message.Data.(map[string]interface{}); ok {
		room, _ := data["room"].(string)
		username, _ := data["username"].(string)
		messageText, _ := data["message"].(string)

		if room != "" {
			log.Printf("💬 Broadcasting chat message from %s to room: %s", username, room)

			// Join room if not already joined
			hub := decorators.GetWebSocketHub()
			if hub != nil {
				hub.JoinGroup(conn.ID, room)

				// Create formatted chat message
				chatMsg := &decorators.WebSocketMessage{
					Type: "chat",
					Data: map[string]interface{}{
						"id":        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
						"user_id":   conn.UserID,
						"username":  username,
						"room":      room,
						"message":   messageText,
						"timestamp": time.Now(),
					},
					Group:     room,
					Timestamp: time.Now(),
				}

				// Broadcast to room
				hub.Broadcast(chatMsg)
			}
		}
	}

	return nil
}

// HandleNotificationMessage handles notification WebSocket messages
// @WebSocket("notification")
func HandleNotificationMessage(conn *decorators.WebSocketConnection, message *decorators.WebSocketMessage) error {
	log.Printf("🔔 Notification message received: %+v", message.Data)

	// Process notification and broadcast to subscribers
	hub := decorators.GetWebSocketHub()
	if hub != nil {
		// Create formatted notification
		notification := &decorators.WebSocketMessage{
			Type: "notification",
			Data: map[string]interface{}{
				"id":        fmt.Sprintf("notif_%d", time.Now().UnixNano()),
				"title":     "New Notification",
				"body":      "You have a new message",
				"timestamp": time.Now(),
			},
			Group:     "notifications",
			Timestamp: time.Now(),
		}

		hub.Broadcast(notification)
	}

	return nil
}

// HandlePresenceMessage handles user presence WebSocket messages
// @WebSocket("presence")
func HandlePresenceMessage(conn *decorators.WebSocketConnection, message *decorators.WebSocketMessage) error {
	log.Printf("👤 Presence update received: %+v", message.Data)

	// Update user presence and notify others
	if data, ok := message.Data.(map[string]interface{}); ok {
		userID, _ := data["user_id"].(string)
		status, _ := data["status"].(string)

		if userID != "" {
			// Update presence in storage
			activeUsers[userID] = &PresenceInfo{
				UserID:   userID,
				Status:   status,
				LastSeen: time.Now(),
			}

			// Broadcast presence update
			hub := decorators.GetWebSocketHub()
			if hub != nil {
				presenceMsg := &decorators.WebSocketMessage{
					Type: "presence",
					Data: map[string]interface{}{
						"user_id":   userID,
						"status":    status,
						"timestamp": time.Now(),
					},
					Timestamp: time.Now(),
				}

				hub.Broadcast(presenceMsg)
			}
		}
	}

	return nil
}

// HandleLiveUpdateMessage handles live data update WebSocket messages
// @WebSocket("live_update")
func HandleLiveUpdateMessage(conn *decorators.WebSocketConnection, message *decorators.WebSocketMessage) error {
	log.Printf("📊 Live update received: %+v", message.Data)

	// Process and broadcast live data updates
	hub := decorators.GetWebSocketHub()
	if hub != nil {
		// Create formatted live update
		liveUpdate := &decorators.WebSocketMessage{
			Type: "live_update",
			Data: map[string]interface{}{
				"type":      "data_update",
				"resource":  "dashboard",
				"data":      message.Data,
				"timestamp": time.Now(),
			},
			Group:     "live_updates",
			Timestamp: time.Now(),
		}

		hub.Broadcast(liveUpdate)
	}

	return nil
}
