package decorators

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketUpgrader configuration for connection upgrade WebSocket
var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// By default, accept all origins (configurable)
		return true
	},
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *WebSocketHub
	UserID   string
	Groups   map[string]bool
	Metadata map[string]interface{}
	mu       sync.RWMutex
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	// Connections ativas
	connections map[string]*WebSocketConnection

	// Grupos de connections
	groups map[string]map[string]*WebSocketConnection

	// Channel for broadcast
	broadcast chan *WebSocketMessage

	// Channel to register connections
	register chan *WebSocketConnection

	// Channel to unregister connections
	unregister chan *WebSocketConnection

	// Mutex for thread safety
	mu sync.RWMutex

	// Configuration
	config WebSocketConfig
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      interface{}            `json:"data"`
	Sender    string                 `json:"sender,omitempty"`
	Target    string                 `json:"target,omitempty"` // ID da specific connection
	Group     string                 `json:"group,omitempty"`  // Nome do grupo
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WebSocketHandler handler type for WebSocket messages
type WebSocketHandler func(conn *WebSocketConnection, message *WebSocketMessage) error

// WebSocketRouter router for WebSocket messages
type WebSocketRouter struct {
	handlers map[string]WebSocketHandler
	mu       sync.RWMutex
}

// Default global hub
var defaultHub *WebSocketHub
var defaultRouter *WebSocketRouter

// InitWebSocket initializes the WebSocket system
func InitWebSocket(config WebSocketConfig) *WebSocketHub {
	// Configure upgrader
	WebSocketUpgrader.ReadBufferSize = config.ReadBuffer
	WebSocketUpgrader.WriteBufferSize = config.WriteBuffer
	WebSocketUpgrader.CheckOrigin = func(r *http.Request) bool {
		return !config.CheckOrigin // If CheckOrigin is false, accept all origins
	}

	hub := &WebSocketHub{
		connections: make(map[string]*WebSocketConnection),
		groups:      make(map[string]map[string]*WebSocketConnection),
		broadcast:   make(chan *WebSocketMessage, 256),
		register:    make(chan *WebSocketConnection),
		unregister:  make(chan *WebSocketConnection),
		config:      config,
	}

	// Start router
	defaultRouter = &WebSocketRouter{
		handlers: make(map[string]WebSocketHandler),
	}

	// Start hub goroutine
	go hub.run()

	// Register default handlers
	RegisterDefaultHandlers()

	defaultHub = hub
	return hub
}

// run executes the main hub loop
func (h *WebSocketHub) run() {
	ticker := time.NewTicker(54 * time.Second) // Ping interval
	defer ticker.Stop()

	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)

		case conn := <-h.unregister:
			h.unregisterConnection(conn)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			h.pingConnections()
		}
	}
}

// registerConnection registers a new connection
func (h *WebSocketHub) registerConnection(conn *WebSocketConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[conn.ID] = conn
	log.Printf("WebSocket: New connection registered %s", conn.ID)

	// Send welcome message
	welcome := &WebSocketMessage{
		Type:      "welcome",
		Data:      map[string]string{"connection_id": conn.ID},
		Timestamp: time.Now(),
	}
	conn.Send <- []byte(welcome.ToJSON())
}

// unregisterConnection removes a connection
func (h *WebSocketHub) unregisterConnection(conn *WebSocketConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.connections[conn.ID]; exists {
		// Remove from all groups
		for groupName := range conn.Groups {
			h.leaveGroupUnsafe(conn, groupName)
		}

		delete(h.connections, conn.ID)
		close(conn.Send)
		log.Printf("WebSocket: Connection removida %s", conn.ID)
	}
}

// broadcastMessage sends message to recipients
func (h *WebSocketHub) broadcastMessage(message *WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data := []byte(message.ToJSON())

	// Envio directed to specific connection
	if message.Target != "" {
		if conn, exists := h.connections[message.Target]; exists {
			select {
			case conn.Send <- data:
			default:
				h.unregisterConnection(conn)
			}
		}
		return
	}

	// Send to specific group
	if message.Group != "" {
		if group, exists := h.groups[message.Group]; exists {
			for _, conn := range group {
				select {
				case conn.Send <- data:
				default:
					h.unregisterConnection(conn)
				}
			}
		}
		return
	}

	// Broadcast to all connections
	for id, conn := range h.connections {
		select {
		case conn.Send <- data:
		default:
			delete(h.connections, id)
			close(conn.Send)
		}
	}
}

// pingConnections sends ping to all connections
func (h *WebSocketHub) pingConnections() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for id, conn := range h.connections {
		conn.mu.Lock()
		if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Printf("WebSocket: Error ao enviar ping para %s: %v", id, err)
			conn.mu.Unlock()
			h.unregister <- conn
			continue
		}
		conn.mu.Unlock()
	}
}

// JoinGroup adds connection to a group
func (h *WebSocketHub) JoinGroup(connID, groupName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn, exists := h.connections[connID]
	if !exists {
		return fmt.Errorf("connection %s not found", connID)
	}

	if h.groups[groupName] == nil {
		h.groups[groupName] = make(map[string]*WebSocketConnection)
	}

	h.groups[groupName][connID] = conn
	conn.Groups[groupName] = true

	log.Printf("WebSocket: Connection %s entrou no grupo %s", connID, groupName)
	return nil
}

// LeaveGroup removes connection from a group
func (h *WebSocketHub) LeaveGroup(connID, groupName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn, exists := h.connections[connID]
	if !exists {
		return fmt.Errorf("connection %s not found", connID)
	}

	h.leaveGroupUnsafe(conn, groupName)
	return nil
}

// leaveGroupUnsafe removes connection from group (without lock)
func (h *WebSocketHub) leaveGroupUnsafe(conn *WebSocketConnection, groupName string) {
	if group, exists := h.groups[groupName]; exists {
		delete(group, conn.ID)
		delete(conn.Groups, groupName)

		// Remove group if empty
		if len(group) == 0 {
			delete(h.groups, groupName)
		}

		log.Printf("WebSocket: Connection %s saiu do grupo %s", conn.ID, groupName)
	}
}

// Broadcast sends message to all connections
func (h *WebSocketHub) Broadcast(message *WebSocketMessage) {
	message.Timestamp = time.Now()
	h.broadcast <- message
}

// SendToConnection sends message to specific connection
func (h *WebSocketHub) SendToConnection(connID string, message *WebSocketMessage) {
	message.Target = connID
	message.Timestamp = time.Now()
	h.broadcast <- message
}

// SendToGroup sends message to group
func (h *WebSocketHub) SendToGroup(groupName string, message *WebSocketMessage) {
	message.Group = groupName
	message.Timestamp = time.Now()
	h.broadcast <- message
}

// ToJSON converts message to JSON
func (m *WebSocketMessage) ToJSON() string {
	data, _ := json.Marshal(m)
	return string(data)
}

// CreateWebSocketHandler creates handler for WebSocket connections
func CreateWebSocketHandler(config *WebSocketConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "WebSocket not habilitado",
			})
		}
	}

	// Initialize hub if it does not exist
	if defaultHub == nil {
		InitWebSocket(*config)
	}

	return func(c *gin.Context) {
		// Upgrade to WebSocket
		conn, err := WebSocketUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket: Error no upgrade: %v", err)
			return
		}

		// Create connection
		wsConn := &WebSocketConnection{
			ID:       generateConnectionID(),
			Conn:     conn,
			Send:     make(chan []byte, 256),
			Hub:      defaultHub,
			UserID:   c.GetString("user_id"), // Get from context if authenticated
			Groups:   make(map[string]bool),
			Metadata: make(map[string]interface{}),
		}

		// Register connection
		defaultHub.register <- wsConn

		// Start goroutines
		go wsConn.writePump()
		go wsConn.readPump()
	}
}

// readPump processes received messages
func (c *WebSocketConnection) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	// Configure timeouts
	pongTimeout, _ := time.ParseDuration(c.Hub.config.PongTimeout)
	c.Conn.SetReadDeadline(time.Now().Add(pongTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket: Error: %v", err)
			}
			break
		}

		// Parse the message
		var message WebSocketMessage
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("WebSocket: Error ao fazer parse da mensagem: %v", err)
			continue
		}

		message.Sender = c.ID
		message.Timestamp = time.Now()

		// Process with router
		if defaultRouter != nil {
			defaultRouter.HandleMessage(c, &message)
		}
	}
}

// writePump sends messages to client
func (c *WebSocketConnection) writePump() {
	pingInterval, _ := time.ParseDuration(c.Hub.config.PingInterval)
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
			if err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// RegisterHandler registers handler for message type
func (r *WebSocketRouter) RegisterHandler(messageType string, handler WebSocketHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[messageType] = handler
}

// HandleMessage processes message using registered handlers
func (r *WebSocketRouter) HandleMessage(conn *WebSocketConnection, message *WebSocketMessage) {
	r.mu.RLock()
	handler, exists := r.handlers[message.Type]
	r.mu.RUnlock()

	if exists {
		if err := handler(conn, message); err != nil {
			log.Printf("WebSocket: Handler error %s: %v", message.Type, err)
		}
	} else {
		log.Printf("WebSocket: Handler not found for type %s", message.Type)
	}
}

// generateConnectionID generates unique ID for connection
func generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

// Default handlers

// JoinGroupHandler handler to join group
func JoinGroupHandler(conn *WebSocketConnection, message *WebSocketMessage) error {
	if data, ok := message.Data.(map[string]interface{}); ok {
		if groupName, ok := data["group"].(string); ok {
			return conn.Hub.JoinGroup(conn.ID, groupName)
		}
	}
	return fmt.Errorf("grupo not especificado")
}

// LeaveGroupHandler handler to leave group
func LeaveGroupHandler(conn *WebSocketConnection, message *WebSocketMessage) error {
	if data, ok := message.Data.(map[string]interface{}); ok {
		if groupName, ok := data["group"].(string); ok {
			return conn.Hub.LeaveGroup(conn.ID, groupName)
		}
	}
	return fmt.Errorf("grupo not especificado")
}

// EchoHandler echo handler for testing
func EchoHandler(conn *WebSocketConnection, message *WebSocketMessage) error {
	response := &WebSocketMessage{
		Type:      "echo",
		Data:      message.Data,
		Timestamp: time.Now(),
	}
	conn.Send <- []byte(response.ToJSON())
	return nil
}

// BroadcastHandler handler for broadcast
func BroadcastHandler(conn *WebSocketConnection, message *WebSocketMessage) error {
	message.Sender = conn.ID
	conn.Hub.Broadcast(message)
	return nil
}

// RegisterDefaultHandlers registers default handlers
func RegisterDefaultHandlers() {
	if defaultRouter == nil {
		defaultRouter = &WebSocketRouter{
			handlers: make(map[string]WebSocketHandler),
		}
	}

	defaultRouter.RegisterHandler("join_group", JoinGroupHandler)
	defaultRouter.RegisterHandler("leave_group", LeaveGroupHandler)
	defaultRouter.RegisterHandler("echo", EchoHandler)
	defaultRouter.RegisterHandler("broadcast", BroadcastHandler)
}

// RegisterDefaultWebSocketHandlers is a public alias for RegisterDefaultHandlers
func RegisterDefaultWebSocketHandlers() {
	RegisterDefaultHandlers()
}

// RegisterWebSocketHandler allows applications to register custom WebSocket handlers
func RegisterWebSocketHandler(messageType string, handler WebSocketHandler) {
	if defaultRouter == nil {
		defaultRouter = &WebSocketRouter{
			handlers: make(map[string]WebSocketHandler),
		}
	}
	defaultRouter.RegisterHandler(messageType, handler)
}

// GetWebSocketHub returns the default WebSocket hub for direct access
func GetWebSocketHub() *WebSocketHub {
	return defaultHub
}

// GetWebSocketInfo returns information about WebSocket
func GetWebSocketInfo(config WebSocketConfig) map[string]interface{} {
	info := map[string]interface{}{
		"enabled":       config.Enabled,
		"read_buffer":   config.ReadBuffer,
		"write_buffer":  config.WriteBuffer,
		"check_origin":  config.CheckOrigin,
		"compression":   config.Compression,
		"ping_interval": config.PingInterval,
		"pong_timeout":  config.PongTimeout,
	}

	if defaultHub != nil {
		defaultHub.mu.RLock()
		info["active_connections"] = len(defaultHub.connections)
		info["active_groups"] = len(defaultHub.groups)
		defaultHub.mu.RUnlock()
	}

	return info
}

// WebSocketStatsHandler handler for WebSocket statistics
func WebSocketStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if defaultHub == nil {
			c.JSON(http.StatusOK, gin.H{
				"websocket": "not_initialized",
			})
			return
		}

		defaultHub.mu.RLock()
		stats := map[string]interface{}{
			"active_connections": len(defaultHub.connections),
			"active_groups":      len(defaultHub.groups),
			"groups":             make(map[string]int),
		}

		for groupName, group := range defaultHub.groups {
			stats["groups"].(map[string]int)[groupName] = len(group)
		}
		defaultHub.mu.RUnlock()

		c.JSON(http.StatusOK, gin.H{
			"websocket_stats": stats,
		})
	}
}
