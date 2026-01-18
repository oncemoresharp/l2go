package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/frostwind/l2go/client"
)

// Manager implements the ClientManager interface
type Manager struct {
	clients      map[string]client.GameClient
	config       *client.ManagerConfig
	metrics      *client.ConnectionMetrics
	eventBus     *client.EventBus
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	isShutdown   bool
}

// NewManager creates a new client manager
func NewManager(config *client.ManagerConfig) *Manager {
	if config == nil {
		config = &client.ManagerConfig{
			MaxClients:      100,
			ConnectInterval: 100 * time.Millisecond,
			HealthCheck:     5 * time.Second,
			RetryAttempts:   3,
			RetryDelay:      1 * time.Second,
		}
	}

	manager := &Manager{
		clients:      make(map[string]client.GameClient),
		config:       config,
		metrics:      &client.ConnectionMetrics{},
		eventBus:     client.NewEventBus(),
		shutdownChan: make(chan struct{}),
	}

	// Start health check routine
	manager.startHealthCheck()

	return manager
}

// Start starts the manager and its background routines
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isShutdown {
		return client.ErrClientManagerClosed
	}

	// Start health check routine
	m.startHealthCheck()

	return nil
}

// CreateClients creates the specified number of clients with the given configuration
func (m *Manager) CreateClients(count int, config client.ClientConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isShutdown {
		return client.ErrClientManagerClosed
	}

	// Check if we would exceed the maximum number of clients
	if len(m.clients)+count > m.config.MaxClients {
		return client.ErrMaxClientsReached
	}

	// Validate the client configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid client configuration: %w", err)
	}

	// Create clients
	for i := 0; i < count; i++ {
		clientID := fmt.Sprintf("client-%d-%d", time.Now().Unix(), i)

		// Check if client already exists (shouldn't happen with timestamp-based IDs)
		if _, exists := m.clients[clientID]; exists {
			return client.ErrClientAlreadyExists
		}

		// Create new client (this would be implemented in the actual GameClient)
		gameClient := NewGameClient(clientID, config)
		m.clients[clientID] = gameClient
	}

	// Update metrics
	m.updateMetrics()

	return nil
}

// StartClients starts the specified clients
func (m *Manager) StartClients(clientIDs []string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.isShutdown {
		return client.ErrClientManagerClosed
	}

	var errors []error

	for _, clientID := range clientIDs {
		gameClient, exists := m.clients[clientID]
		if !exists {
			errors = append(errors, fmt.Errorf("client %s: %w", clientID, client.ErrClientNotFound))
			continue
		}

		// Start client in a goroutine
		m.wg.Add(1)
		go func(id string, gc client.GameClient) {
			defer m.wg.Done()

			if err := gc.Connect(); err != nil {
				m.eventBus.Publish("client.error", map[string]interface{}{
					"clientID": id,
					"error":    err,
					"action":   "connect",
				})
			} else {
				m.eventBus.Publish("client.connected", map[string]interface{}{
					"clientID": id,
				})
			}
		}(clientID, gameClient)

		// Add delay between connections if configured
		if m.config.ConnectInterval > 0 {
			time.Sleep(m.config.ConnectInterval)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start some clients: %v", errors)
	}

	return nil
}

// StopClients stops the specified clients
func (m *Manager) StopClients(clientIDs []string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.isShutdown {
		return client.ErrClientManagerClosed
	}

	var errors []error

	for _, clientID := range clientIDs {
		gameClient, exists := m.clients[clientID]
		if !exists {
			errors = append(errors, fmt.Errorf("client %s: %w", clientID, client.ErrClientNotFound))
			continue
		}

		// Stop client
		if err := gameClient.Disconnect(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop client %s: %w", clientID, err))
		} else {
			m.eventBus.Publish("client.disconnected", map[string]interface{}{
				"clientID": clientID,
			})
		}
	}

	// Update metrics
	m.updateMetrics()

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some clients: %v", errors)
	}

	return nil
}

// GetClient retrieves a client by ID
func (m *Manager) GetClient(clientID string) (client.GameClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gameClient, exists := m.clients[clientID]
	if !exists {
		return nil, client.ErrClientNotFound
	}

	return gameClient, nil
}

// GetAllClients returns all managed clients
func (m *Manager) GetAllClients() map[string]client.GameClient {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	clients := make(map[string]client.GameClient)
	for id, gameClient := range m.clients {
		clients[id] = gameClient
	}

	return clients
}

// GetMetrics returns connection metrics
func (m *Manager) GetMetrics() *client.ConnectionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &client.ConnectionMetrics{
		TotalConnections:   m.metrics.TotalConnections,
		ActiveConnections:  m.metrics.ActiveConnections,
		FailedConnections:  m.metrics.FailedConnections,
		AverageConnectTime: m.metrics.AverageConnectTime,
		LastUpdateTime:     m.metrics.LastUpdateTime,
	}
}

// GetClientStatus returns the status of a specific client
func (m *Manager) GetClientStatus(clientID string) (*client.ClientStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gameClient, exists := m.clients[clientID]
	if !exists {
		return nil, client.ErrClientNotFound
	}

	// Get client status (this would be implemented in the actual GameClient)
	status := &client.ClientStatus{
		ID:            clientID,
		State:         gameClient.GetState(),
		ConnectedTime: time.Now(), // This would be tracked by the actual client
		LastActivity:  time.Now(), // This would be tracked by the actual client
		ErrorCount:    0,          // This would be tracked by the actual client
		LastError:     "",         // This would be tracked by the actual client
	}

	return status, nil
}

// Shutdown gracefully shuts down all clients and the manager
func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isShutdown {
		return nil
	}

	m.isShutdown = true
	close(m.shutdownChan)

	// Stop all clients
	var errors []error
	for clientID, gameClient := range m.clients {
		if err := gameClient.Disconnect(); err != nil {
			errors = append(errors, fmt.Errorf("failed to disconnect client %s: %w", clientID, err))
		}
	}

	// Wait for all goroutines to finish
	m.wg.Wait()

	// Clear clients map
	m.clients = make(map[string]client.GameClient)

	// Update metrics
	m.updateMetrics()

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	return nil
}

// updateMetrics updates the connection metrics
func (m *Manager) updateMetrics() {
	var active, failed int64
	total := int64(len(m.clients))

	for _, gameClient := range m.clients {
		state := gameClient.GetState()
		switch state {
		case client.StateInGame, client.StateConnectingLogin, client.StateAuthenticating, client.StateSelectingServer, client.StateConnectingGame:
			active++
		case client.StateError:
			failed++
		}
	}

	m.metrics.Update(total, active, failed, 0) // AverageConnectTime would be calculated from actual connection times
}

// startHealthCheck starts the health check routine
func (m *Manager) startHealthCheck() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(m.config.HealthCheck)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.performHealthCheck()
			case <-m.shutdownChan:
				return
			}
		}
	}()
}

// performHealthCheck performs health checks on all clients
func (m *Manager) performHealthCheck() {
	m.mu.RLock()
	clients := make(map[string]client.GameClient)
	for id, gameClient := range m.clients {
		clients[id] = gameClient
	}
	m.mu.RUnlock()

	for clientID, gameClient := range clients {
		state := gameClient.GetState()
		if state == client.StateError {
			m.eventBus.Publish("client.health.error", map[string]interface{}{
				"clientID": clientID,
				"state":    state,
			})
		}
	}

	// Update metrics after health check
	m.mu.Lock()
	m.updateMetrics()
	m.mu.Unlock()
}

// NewGameClient creates a new game client (placeholder implementation)
// This would be replaced with the actual GameClient implementation
func NewGameClient(id string, config client.ClientConfig) client.GameClient {
	return &MockGameClient{
		id:     id,
		config: config,
		state:  client.StateDisconnected,
	}
}

// MockGameClient is a placeholder implementation for testing
type MockGameClient struct {
	id     string
	config client.ClientConfig
	state  client.ClientState
	mu     sync.RWMutex
}

func (m *MockGameClient) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = client.StateInGame
	return nil
}

func (m *MockGameClient) Login(username, password string) error {
	return nil
}

func (m *MockGameClient) SelectServer(serverID int) error {
	return nil
}

func (m *MockGameClient) ConnectToGame() error {
	return nil
}

func (m *MockGameClient) CreateCharacter(name string, template *client.CharacterTemplate) error {
	return nil
}

func (m *MockGameClient) SelectCharacter(characterID int) error {
	return nil
}

func (m *MockGameClient) GetCharacterList() ([]client.CharacterInfo, error) {
	return nil, nil
}

func (m *MockGameClient) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = client.StateDisconnected
	return nil
}

func (m *MockGameClient) GetState() client.ClientState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *MockGameClient) GetID() string {
	return m.id
}
