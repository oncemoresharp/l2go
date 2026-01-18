package client

import "errors"

// Configuration validation errors
var (
	ErrInvalidLoginServerHost = errors.New("invalid login server host: must not be empty")
	ErrInvalidLoginServerPort = errors.New("invalid login server port: must be between 1 and 65535")
	ErrInvalidGameServerHost  = errors.New("invalid game server host: must not be empty")
	ErrInvalidGameServerPort  = errors.New("invalid game server port: must be between 1 and 65535")
	ErrInvalidUsername        = errors.New("invalid username: must not be empty")
	ErrInvalidPassword        = errors.New("invalid password: must not be empty")
	ErrInvalidTimeout         = errors.New("invalid timeout: must be greater than 0")
)

// Connection errors
var (
	ErrConnectionFailed  = errors.New("failed to establish connection")
	ErrConnectionTimeout = errors.New("connection timeout")
	ErrConnectionClosed  = errors.New("connection is closed")
	ErrAlreadyConnected  = errors.New("already connected")
	ErrNotConnected      = errors.New("not connected")
)

// Authentication errors
var (
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountBanned        = errors.New("account is banned")
	ErrServerFull           = errors.New("server is full")
)

// Protocol errors
var (
	ErrInvalidPacket     = errors.New("invalid packet format")
	ErrPacketTooLarge    = errors.New("packet too large")
	ErrPacketTooSmall    = errors.New("packet too small")
	ErrUnsupportedOpcode = errors.New("unsupported opcode")
	ErrEncryptionFailed  = errors.New("encryption failed")
	ErrDecryptionFailed  = errors.New("decryption failed")
	ErrChecksumMismatch  = errors.New("checksum mismatch")
)

// Client management errors
var (
	ErrClientNotFound      = errors.New("client not found")
	ErrClientAlreadyExists = errors.New("client already exists")
	ErrMaxClientsReached   = errors.New("maximum number of clients reached")
	ErrClientManagerClosed = errors.New("client manager is closed")
)

// Character management errors
var (
	ErrCharacterNotFound    = errors.New("character not found")
	ErrCharacterNameTaken   = errors.New("character name is already taken")
	ErrInvalidCharacterName = errors.New("invalid character name")
	ErrMaxCharactersReached = errors.New("maximum number of characters reached")
)

// Session errors
var (
	ErrSessionExpired   = errors.New("session expired")
	ErrInvalidSession   = errors.New("invalid session")
	ErrSessionNotFound  = errors.New("session not found")
	ErrMultipleSessions = errors.New("multiple sessions not allowed")
)

// General errors
var (
	ErrInvalidState      = errors.New("invalid client state")
	ErrOperationTimeout  = errors.New("operation timeout")
	ErrResourceExhausted = errors.New("resource exhausted")
	ErrInternalError     = errors.New("internal error")
)
