package client

import (
	"net"
)

// GameClient represents the main client interface for connecting to L2Go servers
type GameClient interface {
	// Connect initiates the full connection sequence (login -> server selection -> game)
	Connect() error

	// Login authenticates with the login server
	Login(username, password string) error

	// SelectServer selects a game server from the available list
	SelectServer(serverID int) error

	// ConnectToGame connects to the selected game server
	ConnectToGame() error

	// CreateCharacter creates a new character with the given template
	CreateCharacter(name string, template *CharacterTemplate) error

	// SelectCharacter selects an existing character
	SelectCharacter(characterID int) error

	// GetCharacterList retrieves the list of characters for the account
	GetCharacterList() ([]CharacterInfo, error)

	// Disconnect gracefully disconnects from all servers
	Disconnect() error

	// GetState returns the current client state
	GetState() ClientState

	// GetID returns the unique client identifier
	GetID() string
}

// ProtocolHandler manages packet encoding/decoding and protocol operations
type ProtocolHandler interface {
	// EncodeLoginPacket encodes a packet for the login server
	EncodeLoginPacket(opcode byte, data []byte) ([]byte, error)

	// DecodeLoginPacket decodes a packet from the login server
	DecodeLoginPacket(raw []byte) (opcode byte, data []byte, err error)

	// EncodeGamePacket encodes a packet for the game server
	EncodeGamePacket(opcode byte, data []byte) ([]byte, error)

	// DecodeGamePacket decodes a packet from the game server
	DecodeGamePacket(raw []byte) (opcode byte, data []byte, err error)

	// InitializeBlowfish initializes Blowfish encryption for login server
	InitializeBlowfish(key []byte) error

	// InitializeXOR initializes XOR encryption for game server
	InitializeXOR(key []byte) error
}

// ClientManager manages multiple concurrent client connections
type ClientManager interface {
	// CreateClients creates the specified number of clients with the given configuration
	CreateClients(count int, config ClientConfig) error

	// StartClients starts the specified clients
	StartClients(clientIDs []string) error

	// StopClients stops the specified clients
	StopClients(clientIDs []string) error

	// GetClient retrieves a client by ID
	GetClient(clientID string) (GameClient, error)

	// GetAllClients returns all managed clients
	GetAllClients() map[string]GameClient

	// GetMetrics returns connection metrics
	GetMetrics() *ConnectionMetrics

	// GetClientStatus returns the status of a specific client
	GetClientStatus(clientID string) (*ClientStatus, error)

	// Shutdown gracefully shuts down all clients and the manager
	Shutdown() error
}

// Connection represents a network connection to a server
type Connection interface {
	// Connect establishes the connection
	Connect(host string, port int) error

	// Send sends data over the connection
	Send(data []byte) error

	// Receive receives data from the connection
	Receive() ([]byte, error)

	// Close closes the connection
	Close() error

	// IsConnected returns whether the connection is active
	IsConnected() bool

	// GetConnection returns the underlying net.Conn
	GetConnection() net.Conn
}
