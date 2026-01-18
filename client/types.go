package client

import (
	"net"
	"sync"
	"time"
)

// ClientState represents the current state of a game client
type ClientState int

const (
	StateDisconnected ClientState = iota
	StateConnectingLogin
	StateAuthenticating
	StateSelectingServer
	StateConnectingGame
	StateInGame
	StateError
)

func (s ClientState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnectingLogin:
		return "ConnectingLogin"
	case StateAuthenticating:
		return "Authenticating"
	case StateSelectingServer:
		return "SelectingServer"
	case StateConnectingGame:
		return "ConnectingGame"
	case StateInGame:
		return "InGame"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// ClientConfig holds configuration for a game client
type ClientConfig struct {
	LoginServerHost string        `json:"loginServerHost"`
	LoginServerPort int           `json:"loginServerPort"`
	GameServerHost  string        `json:"gameServerHost"`
	GameServerPort  int           `json:"gameServerPort"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	AutoCreate      bool          `json:"autoCreate"`
	Timeout         time.Duration `json:"timeout"`
}

// Validate validates the client configuration
func (c *ClientConfig) Validate() error {
	if c.LoginServerHost == "" {
		return ErrInvalidLoginServerHost
	}
	if c.LoginServerPort <= 0 || c.LoginServerPort > 65535 {
		return ErrInvalidLoginServerPort
	}
	if c.GameServerHost == "" {
		return ErrInvalidGameServerHost
	}
	if c.GameServerPort <= 0 || c.GameServerPort > 65535 {
		return ErrInvalidGameServerPort
	}
	if c.Username == "" {
		return ErrInvalidUsername
	}
	if c.Password == "" {
		return ErrInvalidPassword
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second // Default timeout
	}
	return nil
}

// ConnectionMetrics holds metrics about client connections
type ConnectionMetrics struct {
	TotalConnections   int64         `json:"totalConnections"`
	ActiveConnections  int64         `json:"activeConnections"`
	FailedConnections  int64         `json:"failedConnections"`
	AverageConnectTime time.Duration `json:"averageConnectTime"`
	LastUpdateTime     time.Time     `json:"lastUpdateTime"`
	mu                 sync.RWMutex
}

// Update updates the metrics in a thread-safe manner
func (m *ConnectionMetrics) Update(total, active, failed int64, avgTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalConnections = total
	m.ActiveConnections = active
	m.FailedConnections = failed
	m.AverageConnectTime = avgTime
	m.LastUpdateTime = time.Now()
}

// GetSnapshot returns a snapshot of the current metrics
func (m *ConnectionMetrics) GetSnapshot() ConnectionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return ConnectionMetrics{
		TotalConnections:   m.TotalConnections,
		ActiveConnections:  m.ActiveConnections,
		FailedConnections:  m.FailedConnections,
		AverageConnectTime: m.AverageConnectTime,
		LastUpdateTime:     m.LastUpdateTime,
	}
}

// ClientStatus represents the status of a client
type ClientStatus struct {
	ID            string      `json:"id"`
	State         ClientState `json:"state"`
	ConnectedTime time.Time   `json:"connectedTime"`
	LastActivity  time.Time   `json:"lastActivity"`
	ErrorCount    int         `json:"errorCount"`
	LastError     string      `json:"lastError"`
}

// CharacterTemplate represents a character creation template
type CharacterTemplate struct {
	Race      int `json:"race"`
	Class     int `json:"class"`
	Gender    int `json:"gender"`
	HairStyle int `json:"hairStyle"`
	HairColor int `json:"hairColor"`
	Face      int `json:"face"`
}

// CharacterInfo represents character information
type CharacterInfo struct {
	ID       int                `json:"id"`
	Name     string             `json:"name"`
	Level    int                `json:"level"`
	Class    int                `json:"class"`
	Race     int                `json:"race"`
	Gender   int                `json:"gender"`
	Location *CharacterLocation `json:"location"`
	Stats    *CharacterStats    `json:"stats"`
}

// CharacterLocation represents a character's location
type CharacterLocation struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// CharacterStats represents character statistics
type CharacterStats struct {
	HP  int `json:"hp"`
	MP  int `json:"mp"`
	EXP int `json:"exp"`
	SP  int `json:"sp"`
}

// LoginConnection represents a connection to the login server
type LoginConnection struct {
	conn        net.Conn
	sessionID   []byte
	isConnected bool
	mu          sync.RWMutex
}

// GameConnection represents a connection to the game server
type GameConnection struct {
	conn        net.Conn
	isConnected bool
	mu          sync.RWMutex
}

// SessionManager manages login and game sessions
type SessionManager struct {
	loginSession *LoginSession
	gameSession  *GameSession
	mu           sync.RWMutex
}

// LoginSession represents a login server session
type LoginSession struct {
	SessionID      []byte       `json:"sessionId"`
	AccountInfo    *AccountInfo `json:"accountInfo"`
	ServerList     []ServerInfo `json:"serverList"`
	SelectedServer *ServerInfo  `json:"selectedServer"`
}

// GameSession represents a game server session
type GameSession struct {
	Characters   []CharacterInfo `json:"characters"`
	SelectedChar *CharacterInfo  `json:"selectedChar"`
	GameState    *GameState      `json:"gameState"`
}

// AccountInfo represents account information
type AccountInfo struct {
	Username    string    `json:"username"`
	AccountID   int       `json:"accountId"`
	AccessLevel int       `json:"accessLevel"`
	LastLogin   time.Time `json:"lastLogin"`
}

// ServerInfo represents game server information
type ServerInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Status     int    `json:"status"`
	Population int    `json:"population"`
	MaxPlayers int    `json:"maxPlayers"`
}

// GameState represents the current game state
type GameState struct {
	IsInGame    bool      `json:"isInGame"`
	LastUpdate  time.Time `json:"lastUpdate"`
	ServerTime  int64     `json:"serverTime"`
	PlayerCount int       `json:"playerCount"`
	WorldStatus string    `json:"worldStatus"`
}

// Packet represents a network packet
type Packet struct {
	Opcode    byte      `json:"opcode"`
	Data      []byte    `json:"data"`
	Length    uint16    `json:"length"`
	Timestamp time.Time `json:"timestamp"`
}

// EventHandler represents an event handler function
type EventHandler func(event interface{}) error

// EventBus manages event distribution
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe adds an event handler for the specified event type
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish publishes an event to all registered handlers
func (eb *EventBus) Publish(eventType string, event interface{}) {
	eb.mu.RLock()
	handlers := eb.handlers[eventType]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event) // Execute handlers concurrently
	}
}
