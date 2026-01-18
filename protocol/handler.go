package protocol

import (
	"crypto/cipher"
	"fmt"
	"sync"

	"github.com/frostwind/l2go/gameserver/crypt/xor"
	"golang.org/x/crypto/blowfish"
)

// Handler implements the ProtocolHandler interface
type Handler struct {
	loginProtocol *LoginProtocol
	gameProtocol  *GameProtocol
	cryptoEngine  *CryptoEngine
	mu            sync.RWMutex
}

// NewHandler creates a new protocol handler
func NewHandler() *Handler {
	return &Handler{
		loginProtocol: NewLoginProtocol(),
		gameProtocol:  NewGameProtocol(),
		cryptoEngine:  NewCryptoEngine(),
	}
}

// EncodeLoginPacket encodes a packet for the login server
func (h *Handler) EncodeLoginPacket(opcode byte, data []byte) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.loginProtocol.EncodePacket(opcode, data, h.cryptoEngine)
}

// DecodeLoginPacket decodes a packet from the login server
func (h *Handler) DecodeLoginPacket(raw []byte) (opcode byte, data []byte, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.loginProtocol.DecodePacket(raw, h.cryptoEngine)
}

// EncodeGamePacket encodes a packet for the game server
func (h *Handler) EncodeGamePacket(opcode byte, data []byte) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.gameProtocol.EncodePacket(opcode, data, h.cryptoEngine)
}

// DecodeGamePacket decodes a packet from the game server
func (h *Handler) DecodeGamePacket(raw []byte) (opcode byte, data []byte, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.gameProtocol.DecodePacket(raw, h.cryptoEngine)
}

// InitializeBlowfish initializes Blowfish encryption for login server
func (h *Handler) InitializeBlowfish(key []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.cryptoEngine.InitializeBlowfish(key)
}

// InitializeXOR initializes XOR encryption for game server
func (h *Handler) InitializeXOR(key []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.cryptoEngine.InitializeXOR(key)
}

// LoginProtocol handles login server protocol operations
type LoginProtocol struct {
	mu sync.RWMutex
}

// NewLoginProtocol creates a new login protocol handler
func NewLoginProtocol() *LoginProtocol {
	return &LoginProtocol{}
}

// EncodePacket encodes a login server packet
func (lp *LoginProtocol) EncodePacket(opcode byte, data []byte, crypto *CryptoEngine) ([]byte, error) {
	// Create packet with opcode and data
	packet := make([]byte, 1+len(data))
	packet[0] = opcode
	copy(packet[1:], data)

	// Encrypt if Blowfish is initialized
	if crypto.HasBlowfish() {
		encrypted, err := crypto.EncryptBlowfish(packet)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt login packet: %w", err)
		}
		return encrypted, nil
	}

	return packet, nil
}

// DecodePacket decodes a login server packet
func (lp *LoginProtocol) DecodePacket(raw []byte, crypto *CryptoEngine) (opcode byte, data []byte, err error) {
	if len(raw) == 0 {
		return 0, nil, fmt.Errorf("empty packet")
	}

	packet := raw

	// Decrypt if Blowfish is initialized
	if crypto.HasBlowfish() {
		decrypted, err := crypto.DecryptBlowfish(raw)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to decrypt login packet: %w", err)
		}
		packet = decrypted
	}

	if len(packet) == 0 {
		return 0, nil, fmt.Errorf("empty decrypted packet")
	}

	opcode = packet[0]
	if len(packet) > 1 {
		data = packet[1:]
	}

	return opcode, data, nil
}

// GameProtocol handles game server protocol operations
type GameProtocol struct {
	mu sync.RWMutex
}

// NewGameProtocol creates a new game protocol handler
func NewGameProtocol() *GameProtocol {
	return &GameProtocol{}
}

// EncodePacket encodes a game server packet
func (gp *GameProtocol) EncodePacket(opcode byte, data []byte, crypto *CryptoEngine) ([]byte, error) {
	// Create packet with opcode and data
	packet := make([]byte, 1+len(data))
	packet[0] = opcode
	copy(packet[1:], data)

	// Encrypt if XOR is initialized
	if crypto.HasXOR() {
		encrypted, err := crypto.EncryptXOR(packet)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt game packet: %w", err)
		}
		return encrypted, nil
	}

	return packet, nil
}

// DecodePacket decodes a game server packet
func (gp *GameProtocol) DecodePacket(raw []byte, crypto *CryptoEngine) (opcode byte, data []byte, err error) {
	if len(raw) == 0 {
		return 0, nil, fmt.Errorf("empty packet")
	}

	packet := raw

	// Decrypt if XOR is initialized
	if crypto.HasXOR() {
		decrypted, err := crypto.DecryptXOR(raw)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to decrypt game packet: %w", err)
		}
		packet = decrypted
	}

	if len(packet) == 0 {
		return 0, nil, fmt.Errorf("empty decrypted packet")
	}

	opcode = packet[0]
	if len(packet) > 1 {
		data = packet[1:]
	}

	return opcode, data, nil
}

// CryptoEngine manages encryption operations
type CryptoEngine struct {
	blowfishCipher cipher.Block
	xorCipher      *xor.Cipher
	mu             sync.RWMutex
}

// NewCryptoEngine creates a new crypto engine
func NewCryptoEngine() *CryptoEngine {
	return &CryptoEngine{}
}

// InitializeBlowfish initializes Blowfish encryption
func (ce *CryptoEngine) InitializeBlowfish(key []byte) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	cipher, err := blowfish.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create Blowfish cipher: %w", err)
	}

	ce.blowfishCipher = cipher
	return nil
}

// InitializeXOR initializes XOR encryption
func (ce *CryptoEngine) InitializeXOR(key []byte) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	cipher := xor.NewCipher()
	// Copy the provided key to the cipher's keys
	if len(key) >= 8 {
		copy(cipher.InputKey, key[:8])
		copy(cipher.OutputKey, key[:8])
	}
	ce.xorCipher = cipher
	return nil
}

// HasBlowfish returns true if Blowfish encryption is initialized
func (ce *CryptoEngine) HasBlowfish() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.blowfishCipher != nil
}

// HasXOR returns true if XOR encryption is initialized
func (ce *CryptoEngine) HasXOR() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.xorCipher != nil
}

// EncryptBlowfish encrypts data using Blowfish
func (ce *CryptoEngine) EncryptBlowfish(data []byte) ([]byte, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if ce.blowfishCipher == nil {
		return nil, fmt.Errorf("Blowfish cipher not initialized")
	}

	// Pad data to block size
	blockSize := ce.blowfishCipher.BlockSize()
	padded := make([]byte, ((len(data)+blockSize-1)/blockSize)*blockSize)
	copy(padded, data)

	// Encrypt in blocks
	encrypted := make([]byte, len(padded))
	for i := 0; i < len(padded); i += blockSize {
		ce.blowfishCipher.Encrypt(encrypted[i:i+blockSize], padded[i:i+blockSize])
	}

	return encrypted, nil
}

// DecryptBlowfish decrypts data using Blowfish
func (ce *CryptoEngine) DecryptBlowfish(data []byte) ([]byte, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if ce.blowfishCipher == nil {
		return nil, fmt.Errorf("Blowfish cipher not initialized")
	}

	blockSize := ce.blowfishCipher.BlockSize()
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("data length must be multiple of block size")
	}

	// Decrypt in blocks
	decrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += blockSize {
		ce.blowfishCipher.Decrypt(decrypted[i:i+blockSize], data[i:i+blockSize])
	}

	return decrypted, nil
}

// EncryptXOR encrypts data using XOR
func (ce *CryptoEngine) EncryptXOR(data []byte) ([]byte, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if ce.xorCipher == nil {
		return nil, fmt.Errorf("XOR cipher not initialized")
	}

	encrypted := make([]byte, len(data))
	copy(encrypted, data)
	
	// Make a copy of the output key for encryption
	key := make([]byte, len(ce.xorCipher.OutputKey))
	copy(key, ce.xorCipher.OutputKey)
	
	xor.Encrypt(encrypted, key)
	return encrypted, nil
}

// DecryptXOR decrypts data using XOR
func (ce *CryptoEngine) DecryptXOR(data []byte) ([]byte, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if ce.xorCipher == nil {
		return nil, fmt.Errorf("XOR cipher not initialized")
	}

	decrypted := make([]byte, len(data))
	copy(decrypted, data)
	
	// Make a copy of the input key for decryption
	key := make([]byte, len(ce.xorCipher.InputKey))
	copy(key, ce.xorCipher.InputKey)
	
	xor.Decrypt(decrypted, key)
	return decrypted, nil
}
