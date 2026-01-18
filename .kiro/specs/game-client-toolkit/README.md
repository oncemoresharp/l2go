# Game Client Toolkit

A comprehensive toolkit for connecting multiple clients to the L2Go server infrastructure.

## Project Structure

```
client/
├── interfaces.go      # Core interfaces for GameClient, ProtocolHandler, and ClientManager
├── types.go          # Data types and structures
├── errors.go         # Error definitions
├── config.go         # Configuration system with validation
└── config_test.go    # Configuration tests

protocol/
└── handler.go        # Protocol handler implementation with encryption support

manager/
└── clientmanager.go  # Client manager implementation for multi-client handling

examples/
└── client-toolkit.json  # Example configuration file
```

## Core Components

### GameClient Interface
- Handles complete connection lifecycle (login → server selection → game)
- Supports character management operations
- Provides state management and error handling

### ProtocolHandler Interface
- Manages packet encoding/decoding for both login and game servers
- Supports Blowfish encryption for login server communication
- Supports XOR encryption for game server communication
- Maintains separate encryption contexts

### ClientManager Interface
- Manages multiple concurrent client connections
- Provides connection metrics and monitoring
- Supports graceful shutdown and resource cleanup
- Handles connection limits and failure isolation

## Configuration System

The toolkit supports multiple environment profiles (development, testing, production) with comprehensive validation:

- **Client Configuration**: Server addresses, credentials, timeouts
- **Manager Configuration**: Connection limits, health checks, retry policies
- **Load Test Configuration**: Default test parameters and reporting
- **Logging Configuration**: Log levels, formats, and rotation
- **Environment Profiles**: Environment-specific settings

## Dependencies

- `github.com/leanovate/gopter`: Property-based testing framework
- `golang.org/x/crypto`: Cryptographic functions (Blowfish)
- Existing L2Go XOR cipher implementation

## Usage

```go
// Load configuration
config, err := client.LoadConfig("client-toolkit.json")
if err != nil {
    log.Fatal(err)
}

// Apply active profile
if err := config.ApplyProfile(); err != nil {
    log.Fatal(err)
}

// Create protocol handler
handler := protocol.NewHandler()

// Create client manager
manager := manager.NewManager(&config.Manager)
```

## Requirements Addressed

This implementation addresses the following requirements:

- **7.1**: Configuration file reading with validation
- **7.4**: Support for multiple environment profiles  
- **7.5**: Configuration parameter validation with helpful error messages

The project structure provides a solid foundation for implementing the complete Game Client Toolkit with proper separation of concerns and extensibility.