package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	ErrInsufficientData = errors.New("insufficient data in buffer")
	ErrInvalidString    = errors.New("invalid string format")
	ErrBufferOverflow   = errors.New("buffer overflow")
)

type Buffer struct {
	bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{}
}

func NewBufferFromBytes(data []byte) *Buffer {
	buf := &Buffer{}
	buf.Write(data)
	return buf
}

// Enhanced write methods with error handling
func (b *Buffer) WriteUInt64(value uint64) error {
	return binary.Write(b, binary.LittleEndian, value)
}

func (b *Buffer) WriteUInt32(value uint32) error {
	return binary.Write(b, binary.LittleEndian, value)
}

func (b *Buffer) WriteUInt16(value uint16) error {
	return binary.Write(b, binary.LittleEndian, value)
}

func (b *Buffer) WriteUInt8(value uint8) error {
	return binary.Write(b, binary.LittleEndian, value)
}

func (b *Buffer) WriteFloat64(value float64) error {
	return binary.Write(b, binary.LittleEndian, value)
}

func (b *Buffer) WriteFloat32(value float32) error {
	return binary.Write(b, binary.LittleEndian, value)
}

// Additional write methods for client use
func (b *Buffer) WriteString(value string) error {
	// Write string as UTF-16LE with null terminator
	for _, r := range value {
		if err := b.WriteUInt16(uint16(r)); err != nil {
			return err
		}
	}
	// Null terminator
	return b.WriteUInt16(0)
}

func (b *Buffer) WriteBytes(data []byte) error {
	_, err := b.Write(data)
	return err
}

func (b *Buffer) WriteBool(value bool) error {
	if value {
		return b.WriteUInt8(1)
	}
	return b.WriteUInt8(0)
}

// Packet construction helpers
func (b *Buffer) WritePacketHeader(opcode byte, length uint16) error {
	if err := b.WriteUInt16(length); err != nil {
		return err
	}
	return b.WriteUInt8(opcode)
}

func (b *Buffer) PrependLength() error {
	data := b.Bytes()
	length := uint16(len(data))

	// Create new buffer with length prefix
	newBuf := NewBuffer()
	if err := newBuf.WriteUInt16(length); err != nil {
		return err
	}
	if err := newBuf.WriteBytes(data); err != nil {
		return err
	}

	// Replace current buffer content
	b.Reset()
	_, err := b.Write(newBuf.Bytes())
	return err
}

// Validation and utility methods
func (b *Buffer) Size() int {
	return b.Len()
}

func (b *Buffer) IsEmpty() bool {
	return b.Len() == 0
}

func (b *Buffer) Clear() {
	b.Reset()
}

func (b *Buffer) Clone() *Buffer {
	newBuf := NewBuffer()
	newBuf.Write(b.Bytes())
	return newBuf
}

type Reader struct {
	*bytes.Reader
}

func NewReader(buffer []byte) *Reader {
	return &Reader{bytes.NewReader(buffer)}
}

func (r *Reader) ReadBytes(number int) []byte {
	buffer := make([]byte, number)
	n, _ := r.Read(buffer)
	if n < number {
		return []byte{}
	}

	return buffer
}

func (r *Reader) ReadUInt64() uint64 {
	var result uint64

	buffer := make([]byte, 8)
	n, _ := r.Read(buffer)
	if n < 8 {
		return 0
	}

	buf := bytes.NewBuffer(buffer)

	binary.Read(buf, binary.LittleEndian, &result)

	return result
}

func (r *Reader) ReadUInt32() uint32 {
	var result uint32

	buffer := make([]byte, 4)
	n, _ := r.Read(buffer)
	if n < 4 {
		return 0
	}

	buf := bytes.NewBuffer(buffer)

	binary.Read(buf, binary.LittleEndian, &result)

	return result
}

func (r *Reader) ReadUInt16() uint16 {
	var result uint16

	buffer := make([]byte, 2)
	n, _ := r.Read(buffer)
	if n < 2 {
		return 0
	}

	buf := bytes.NewBuffer(buffer)

	binary.Read(buf, binary.LittleEndian, &result)

	return result
}

func (r *Reader) ReadUInt8() uint8 {
	var result uint8

	buffer := make([]byte, 1)
	n, _ := r.Read(buffer)
	if n < 1 {
		return 0
	}

	buf := bytes.NewBuffer(buffer)

	binary.Read(buf, binary.LittleEndian, &result)

	return result
}

func (r *Reader) ReadString() string {
	var result []byte
	var first_byte, second_byte byte

	for {
		first_byte, _ = r.ReadByte()
		second_byte, _ = r.ReadByte()
		if first_byte == 0x00 && second_byte == 0x00 {
			break
		} else {
			result = append(result, first_byte, second_byte)
		}
	}

	return string(result)
}
