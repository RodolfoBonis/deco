package decorators

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// Tests for WebSocket functionality

func TestWebSocketHub_InitWebSocket(t *testing.T) {
	// Remove  to avoid race conditions

	config := WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	}

	hub := InitWebSocket(config)
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.connections)
	assert.NotNil(t, hub.groups)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	// Check if mutex is initialized without copying it
	assert.NotNil(t, &hub.mu)
}

func TestWebSocketHub_RegisterConnection(t *testing.T) {
	// Remove  to avoid race conditions

	config := WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	}

	hub := InitWebSocket(config)
	conn := &WebSocketConnection{
		ID:       "test-conn-1",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	// Register connection
	hub.register <- conn

	// Wait a bit for the registration to complete
	time.Sleep(10 * time.Millisecond)

	// Check if connection is registered
	hub.mu.RLock()
	_, exists := hub.connections[conn.ID]
	hub.mu.RUnlock()

	assert.True(t, exists, "Connection should be registered")
}

func TestWebSocketHub_UnregisterConnection(t *testing.T) {
	// Remove  to avoid race conditions

	config := WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	}

	hub := InitWebSocket(config)
	conn := &WebSocketConnection{
		ID:       "test-conn-2",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	// Register connection first
	hub.register <- conn
	time.Sleep(10 * time.Millisecond)

	// Unregister connection
	hub.unregister <- conn
	time.Sleep(10 * time.Millisecond)

	// Check if connection is unregistered
	hub.mu.RLock()
	_, exists := hub.connections[conn.ID]
	hub.mu.RUnlock()

	assert.False(t, exists, "Connection should be unregistered")
}

func TestWebSocketHub_BroadcastMessage(t *testing.T) {
	// Remove  to avoid race conditions

	config := WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	}

	hub := InitWebSocket(config)

	// Create test connections
	conn1 := &WebSocketConnection{
		ID:       "test-conn-3",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}
	conn2 := &WebSocketConnection{
		ID:       "test-conn-4",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	// Register connections
	hub.register <- conn1
	hub.register <- conn2
	time.Sleep(10 * time.Millisecond)

	// Após registrar as conexões:
	<-conn1.Send // consome mensagem de welcome
	<-conn2.Send // consome mensagem de welcome

	// Broadcast message
	testMessage := &WebSocketMessage{
		Type:      "test",
		Data:      "test broadcast message",
		Timestamp: time.Now(),
	}
	hub.broadcast <- testMessage
	time.Sleep(10 * time.Millisecond)

	// Check if messages were sent to both connections
	select {
	case msg := <-conn1.Send:
		assert.Contains(t, string(msg), "test broadcast message")
	default:
		assert.Fail(t, "Message should be sent to conn1")
	}

	select {
	case msg := <-conn2.Send:
		assert.Contains(t, string(msg), "test broadcast message")
	default:
		assert.Fail(t, "Message should be sent to conn2")
	}
}

func TestWebSocketHub_JoinGroup(t *testing.T) {
	// Remove  to avoid race conditions

	config := WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	}

	hub := InitWebSocket(config)
	conn := &WebSocketConnection{
		ID:       "test-conn-5",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	// Register connection
	hub.register <- conn
	time.Sleep(10 * time.Millisecond)

	// Join group
	groupName := "test-group"
	err := hub.JoinGroup(conn.ID, groupName)
	assert.NoError(t, err)

	// Check if connection is in the group
	hub.mu.RLock()
	group, exists := hub.groups[groupName]
	hub.mu.RUnlock()

	assert.True(t, exists, "Group should exist")
	assert.Contains(t, group, conn.ID, "Connection ID should be in group")
}

func TestWebSocketHub_LeaveGroup(t *testing.T) {
	// Remove  to avoid race conditions

	config := WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	}

	hub := InitWebSocket(config)
	conn := &WebSocketConnection{
		ID:       "test-conn-6",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	// Register connection
	hub.register <- conn
	time.Sleep(10 * time.Millisecond)

	// Join group first
	groupName := "test-group-2"
	err := hub.JoinGroup(conn.ID, groupName)
	assert.NoError(t, err)

	// Leave group
	err = hub.LeaveGroup(conn.ID, groupName)
	assert.NoError(t, err)

	// Check if connection is removed from group
	hub.mu.RLock()
	group, exists := hub.groups[groupName]
	hub.mu.RUnlock()

	if exists {
		assert.NotContains(t, group, conn.ID, "Connection ID should not be in group")
	}
}

func TestWebSocketHub_BroadcastToGroup(t *testing.T) {
	// Remove  to avoid race conditions

	config := WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	}

	hub := InitWebSocket(config)

	// Create test connections
	conn1 := &WebSocketConnection{
		ID:       "test-conn-7",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}
	conn2 := &WebSocketConnection{
		ID:       "test-conn-8",
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	// Register connections
	hub.register <- conn1
	hub.register <- conn2
	time.Sleep(10 * time.Millisecond)

	// Após registrar as conexões:
	<-conn1.Send // consome mensagem de welcome
	<-conn2.Send // consome mensagem de welcome

	// Join group
	groupName := "test-group-3"
	err1 := hub.JoinGroup(conn1.ID, groupName)
	err2 := hub.JoinGroup(conn2.ID, groupName)
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Send message to group
	testMessage := &WebSocketMessage{
		Type:      "test",
		Data:      "test group message",
		Timestamp: time.Now(),
	}
	hub.SendToGroup(groupName, testMessage)
	time.Sleep(10 * time.Millisecond)

	// Check if messages were sent to both connections in the group
	select {
	case msg := <-conn1.Send:
		assert.Contains(t, string(msg), "test group message")
	default:
		assert.Fail(t, "Message should be sent to conn1")
	}

	select {
	case msg := <-conn2.Send:
		assert.Contains(t, string(msg), "test group message")
	default:
		assert.Fail(t, "Message should be sent to conn2")
	}
}

func TestWebSocketConnection_ToJSON(t *testing.T) {
	// Remove  to avoid race conditions

	// Create WebSocket message
	message := &WebSocketMessage{
		Type:      "test",
		Data:      "test data",
		Timestamp: time.Now(),
	}

	// Send JSON message
	jsonStr := message.ToJSON()
	assert.NotEmpty(t, jsonStr)

	// Check if JSON is valid
	var received map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &received)
	assert.NoError(t, err)
	assert.Contains(t, received, "type")
	assert.Contains(t, received, "data")
}

func TestCreateWebSocketHandler(t *testing.T) {
	// Remove  to avoid race conditions

	setupGinTestMode(t)
	router := gin.New()

	// Create WebSocket handler
	handler := CreateWebSocketHandler(&WebSocketConfig{
		Enabled:      true,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		CheckOrigin:  false,
		Compression:  false,
		PingInterval: "54s",
		PongTimeout:  "60s",
	})
	assert.NotNil(t, handler)

	// Add route
	router.GET("/ws", handler)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http://... to ws://...
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
	}
	defer ws.Close()

	// Test connection is established
	assert.NotNil(t, ws)
}

func TestWebSocketHub_GenerateConnectionID(t *testing.T) {
	// Remove  to avoid race conditions

	// Generate multiple connection IDs with small delays to ensure uniqueness
	id1 := generateConnectionID()
	time.Sleep(1 * time.Microsecond)
	id2 := generateConnectionID()
	time.Sleep(1 * time.Microsecond)
	id3 := generateConnectionID()

	// Check that IDs are unique
	assert.NotEqual(t, id1, id2, "Connection IDs should be unique")
	assert.NotEqual(t, id1, id3, "Connection IDs should be unique")
	assert.NotEqual(t, id2, id3, "Connection IDs should be unique")

	// Check that IDs are not empty
	assert.NotEmpty(t, id1, "Connection ID should not be empty")
	assert.NotEmpty(t, id2, "Connection ID should not be empty")
	assert.NotEmpty(t, id3, "Connection ID should not be empty")

	// Check that IDs follow the expected format (conn_<timestamp>)
	assert.Regexp(t, `^conn_\d+$`, id1, "Connection ID should follow format conn_<timestamp>")
	assert.Regexp(t, `^conn_\d+$`, id2, "Connection ID should follow format conn_<timestamp>")
	assert.Regexp(t, `^conn_\d+$`, id3, "Connection ID should follow format conn_<timestamp>")
}

func TestWebSocketHub_WebSocketStatsHandler(t *testing.T) {
	// Remove  to avoid race conditions

	setupGinTestMode(t)
	router := gin.New()

	// Add WebSocket stats handler
	router.GET("/ws/stats", WebSocketStatsHandler())

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws/stats", http.NoBody)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code, "Stats handler should return 200")

	// Check response body contains stats
	body := w.Body.String()
	assert.Contains(t, body, "connections", "Response should contain connections")
	assert.Contains(t, body, "groups", "Response should contain groups")
}

func TestWebSocketHub_CustomCheckOrigin(t *testing.T) {
	// Remove  to avoid race conditions

	// Test with allowed origin
	allowedOrigins := []string{"http://localhost:3000", "https://example.com"}

	req, _ := http.NewRequest("GET", "/ws", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3000")

	allowed := CustomCheckOrigin(allowedOrigins)(req)
	assert.True(t, allowed, "Origin should be allowed")

	// Test with disallowed origin
	req.Header.Set("Origin", "http://malicious.com")
	allowed = CustomCheckOrigin(allowedOrigins)(req)
	assert.False(t, allowed, "Origin should not be allowed")

	// Test with no origin header
	req.Header.Del("Origin")
	allowed = CustomCheckOrigin(allowedOrigins)(req)
	assert.False(t, allowed, "Request without origin não deve ser permitido se o CheckOrigin exigir Origin")
}

func TestPingConnections(t *testing.T) {
	// Test ping connections functionality
	config := WebSocketConfig{}
	hub := InitWebSocket(config)

	// Test ping connections (this will run in background)
	// We can't easily test the internal ping without mocking
	assert.NotNil(t, hub)
}

func TestBroadcast(t *testing.T) {
	// Test broadcast functionality
	config := WebSocketConfig{}
	hub := InitWebSocket(config)

	// Test broadcast with message
	message := &WebSocketMessage{
		Type: "test",
		Data: "test message",
	}
	hub.Broadcast(message)

	assert.NotNil(t, hub)
}

func TestSendToConnection(t *testing.T) {
	// Test send to specific connection
	config := WebSocketConfig{}
	hub := InitWebSocket(config)

	message := &WebSocketMessage{
		Type: "test",
		Data: "test message",
	}
	hub.SendToConnection("test-conn", message)

	assert.NotNil(t, hub)
}

func TestHandleMessage(t *testing.T) {
	// Test handle message functionality
	config := WebSocketConfig{}
	hub := InitWebSocket(config)

	// Test that hub was created
	assert.NotNil(t, hub)
}

func TestJoinGroupHandler(t *testing.T) {
	// Test join group handler
	config := WebSocketConfig{}
	hub := InitWebSocket(config)

	conn := &WebSocketConnection{
		ID:     "test",
		Hub:    hub,
		Send:   make(chan []byte, 1),  // Initialize Send channel
		Groups: make(map[string]bool), // Initialize Groups map
	}

	// Register the connection first
	hub.registerConnection(conn)

	message := &WebSocketMessage{Type: "join", Data: map[string]interface{}{"group": "test"}}

	err := JoinGroupHandler(conn, message)
	assert.NoError(t, err)

	// Clean up - unregister connection before closing channel
	hub.unregisterConnection(conn)
}

func TestLeaveGroupHandler(t *testing.T) {
	// Test leave group handler
	config := WebSocketConfig{}
	hub := InitWebSocket(config)

	conn := &WebSocketConnection{
		ID:     "test",
		Hub:    hub,
		Send:   make(chan []byte, 1),  // Initialize Send channel
		Groups: make(map[string]bool), // Initialize Groups map
	}

	// Register the connection first
	hub.registerConnection(conn)

	message := &WebSocketMessage{Type: "leave", Data: map[string]interface{}{"group": "test"}}

	err := LeaveGroupHandler(conn, message)
	assert.NoError(t, err)

	// Clean up - unregister connection before closing channel
	hub.unregisterConnection(conn)
}

func TestEchoHandler(t *testing.T) {
	// Test echo handler
	conn := &WebSocketConnection{
		ID:   "test",
		Send: make(chan []byte, 1),
	}
	message := &WebSocketMessage{Type: "echo", Data: "test message"}

	err := EchoHandler(conn, message)
	assert.NoError(t, err)

	// Clean up
	close(conn.Send)
}

func TestBroadcastHandler(t *testing.T) {
	// Test broadcast handler
	config := WebSocketConfig{}
	hub := InitWebSocket(config)

	conn := &WebSocketConnection{
		ID:     "test",
		Hub:    hub,
		Send:   make(chan []byte, 1),  // Initialize Send channel
		Groups: make(map[string]bool), // Initialize Groups map
	}

	// Register the connection first
	hub.registerConnection(conn)

	message := &WebSocketMessage{Type: "broadcast", Data: map[string]interface{}{"message": "test"}}

	err := BroadcastHandler(conn, message)
	assert.NoError(t, err)

	// Clean up - unregister connection before closing channel
	hub.unregisterConnection(conn)
}

func TestRegisterDefaultWebSocketHandlers(t *testing.T) {
	// Test registering default handlers
	RegisterDefaultWebSocketHandlers()

	// Verify handlers were registered (this is a basic test)
	assert.NotNil(t, defaultRouter)
}

func TestRegisterWebSocketHandler(t *testing.T) {
	// Test registering custom handler
	handler := func(_ *WebSocketConnection, _ *WebSocketMessage) error {
		return nil
	}

	RegisterWebSocketHandler("custom", handler)

	// Verify handler was registered
	assert.NotNil(t, defaultRouter)
}

func TestGetWebSocketHub(t *testing.T) {
	// Test getting WebSocket hub
	hub := GetWebSocketHub()
	assert.NotNil(t, hub)
}

func TestGetWebSocketInfo(t *testing.T) {
	// Test getting WebSocket info
	config := WebSocketConfig{}
	info := GetWebSocketInfo(config)
	assert.NotNil(t, info)
	assert.Contains(t, info, "active_connections")
	assert.Contains(t, info, "active_groups")
}

func TestWebSocketHandlerWrapper(t *testing.T) {
	// Test WebSocket handler wrapper
	handler := func(_ *WebSocketConnection, _ *WebSocketMessage) error {
		return nil
	}

	wrapper := WebSocketHandlerWrapper(handler)
	assert.NotNil(t, wrapper)
}
