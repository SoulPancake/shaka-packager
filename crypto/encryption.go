// Package crypto provides encryption and decryption functionality for media content
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
)

// Encryptor provides encryption functionality for media content
type Encryptor interface {
	Initialize(params EncryptionParams) error
	Encrypt(data []byte) ([]byte, error)
	GetKeyID() []byte
	GetKey() []byte
	Close() error
}

// Decryptor provides decryption functionality for media content
type Decryptor interface {
	Initialize(params DecryptionParams) error
	Decrypt(data []byte) ([]byte, error)
	Close() error
}

// EncryptionParams contains encryption configuration
type EncryptionParams struct {
	KeyID        []byte
	Key          []byte
	IV           []byte
	Method       EncryptionMethod
	ClearBytes   uint32
	ProtectedBytes uint32
}

// DecryptionParams contains decryption configuration
type DecryptionParams struct {
	KeyID  []byte
	Key    []byte
	Method EncryptionMethod
}

// EncryptionMethod represents the encryption algorithm
type EncryptionMethod int

const (
	EncryptionNone EncryptionMethod = iota
	EncryptionAES128
	EncryptionAES192
	EncryptionAES256
	EncryptionSampleAES
)

func (e EncryptionMethod) String() string {
	switch e {
	case EncryptionAES128:
		return "AES-128"
	case EncryptionAES192:
		return "AES-192"
	case EncryptionAES256:
		return "AES-256"
	case EncryptionSampleAES:
		return "Sample-AES"
	case EncryptionNone:
		return "None"
	default:
		return "Unknown"
	}
}

// AESEncryptor implements AES encryption
type AESEncryptor struct {
	keyID        []byte
	key          []byte
	cipher       cipher.Block
	method       EncryptionMethod
	clearBytes   uint32
	protectedBytes uint32
}

// NewAESEncryptor creates a new AES encryptor
func NewAESEncryptor() *AESEncryptor {
	return &AESEncryptor{}
}

// Initialize sets up the encryptor
func (e *AESEncryptor) Initialize(params EncryptionParams) error {
	if len(params.Key) == 0 {
		return fmt.Errorf("encryption key is required")
	}

	// Validate key size
	switch len(params.Key) {
	case 16:
		e.method = EncryptionAES128
	case 24:
		e.method = EncryptionAES192
	case 32:
		e.method = EncryptionAES256
	default:
		return fmt.Errorf("invalid key size: %d bytes", len(params.Key))
	}

	e.key = make([]byte, len(params.Key))
	copy(e.key, params.Key)

	if len(params.KeyID) > 0 {
		e.keyID = make([]byte, len(params.KeyID))
		copy(e.keyID, params.KeyID)
	} else {
		// Generate key ID from key hash
		hash := sha1.Sum(e.key)
		e.keyID = hash[:16]
	}

	e.clearBytes = params.ClearBytes
	e.protectedBytes = params.ProtectedBytes

	// Create cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}
	e.cipher = block

	return nil
}

// Encrypt encrypts the provided data
func (e *AESEncryptor) Encrypt(data []byte) ([]byte, error) {
	if e.cipher == nil {
		return nil, fmt.Errorf("encryptor not initialized")
	}

	if len(data) == 0 {
		return data, nil
	}

	// Generate random IV for each encryption
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Apply subsample encryption if configured
	if e.clearBytes > 0 || e.protectedBytes > 0 {
		return e.encryptSubsample(data, iv)
	}

	// Full encryption
	return e.encryptFull(data, iv)
}

// encryptFull performs full encryption of the data
func (e *AESEncryptor) encryptFull(data []byte, iv []byte) ([]byte, error) {
	// Pad data to block size
	blockSize := e.cipher.BlockSize()
	padSize := blockSize - (len(data) % blockSize)
	if padSize != blockSize {
		padding := make([]byte, padSize)
		for i := range padding {
			padding[i] = byte(padSize)
		}
		data = append(data, padding...)
	}

	// Encrypt
	mode := cipher.NewCBCEncrypter(e.cipher, iv)
	encrypted := make([]byte, len(data))
	mode.CryptBlocks(encrypted, data)

	// Prepend IV to encrypted data
	result := append(iv, encrypted...)
	return result, nil
}

// encryptSubsample performs subsample encryption
func (e *AESEncryptor) encryptSubsample(data []byte, iv []byte) ([]byte, error) {
	if e.clearBytes == 0 && e.protectedBytes == 0 {
		return e.encryptFull(data, iv)
	}

	result := make([]byte, 0, len(data)+aes.BlockSize)
	result = append(result, iv...) // Prepend IV

	pos := 0
	for pos < len(data) {
		// Clear bytes
		clearEnd := pos + int(e.clearBytes)
		if clearEnd > len(data) {
			clearEnd = len(data)
		}
		if clearEnd > pos {
			result = append(result, data[pos:clearEnd]...)
			pos = clearEnd
		}

		if pos >= len(data) {
			break
		}

		// Protected bytes
		protectedEnd := pos + int(e.protectedBytes)
		if protectedEnd > len(data) {
			protectedEnd = len(data)
		}
		if protectedEnd > pos {
			protectedData := data[pos:protectedEnd]
			
			// Encrypt protected portion
			blockSize := e.cipher.BlockSize()
			if len(protectedData)%blockSize != 0 {
				// Pad to block size
				padSize := blockSize - (len(protectedData) % blockSize)
				padding := make([]byte, padSize)
				protectedData = append(protectedData, padding...)
			}
			
			mode := cipher.NewCBCEncrypter(e.cipher, iv)
			encrypted := make([]byte, len(protectedData))
			mode.CryptBlocks(encrypted, protectedData)
			
			result = append(result, encrypted...)
			pos = protectedEnd
		}
	}

	return result, nil
}

// GetKeyID returns the key identifier
func (e *AESEncryptor) GetKeyID() []byte {
	return e.keyID
}

// GetKey returns the encryption key
func (e *AESEncryptor) GetKey() []byte {
	return e.key
}

// Close cleans up the encryptor
func (e *AESEncryptor) Close() error {
	// Clear sensitive data
	if e.key != nil {
		for i := range e.key {
			e.key[i] = 0
		}
		e.key = nil
	}
	e.cipher = nil
	return nil
}

// AESDecryptor implements AES decryption
type AESDecryptor struct {
	key    []byte
	cipher cipher.Block
	method EncryptionMethod
}

// NewAESDecryptor creates a new AES decryptor
func NewAESDecryptor() *AESDecryptor {
	return &AESDecryptor{}
}

// Initialize sets up the decryptor
func (d *AESDecryptor) Initialize(params DecryptionParams) error {
	if len(params.Key) == 0 {
		return fmt.Errorf("decryption key is required")
	}

	d.key = make([]byte, len(params.Key))
	copy(d.key, params.Key)

	// Create cipher
	block, err := aes.NewCipher(d.key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}
	d.cipher = block
	d.method = params.Method

	return nil
}

// Decrypt decrypts the provided data
func (d *AESDecryptor) Decrypt(data []byte) ([]byte, error) {
	if d.cipher == nil {
		return nil, fmt.Errorf("decryptor not initialized")
	}

	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract IV from the beginning
	iv := data[:aes.BlockSize]
	encrypted := data[aes.BlockSize:]

	if len(encrypted) == 0 {
		return []byte{}, nil
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(d.cipher, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Remove padding
	if len(decrypted) > 0 {
		padSize := int(decrypted[len(decrypted)-1])
		if padSize > 0 && padSize <= aes.BlockSize && padSize <= len(decrypted) {
			decrypted = decrypted[:len(decrypted)-padSize]
		}
	}

	return decrypted, nil
}

// Close cleans up the decryptor
func (d *AESDecryptor) Close() error {
	// Clear sensitive data
	if d.key != nil {
		for i := range d.key {
			d.key[i] = 0
		}
		d.key = nil
	}
	d.cipher = nil
	return nil
}

// NewEncryptor creates an encryptor based on the method
func NewEncryptor(method EncryptionMethod) Encryptor {
	switch method {
	case EncryptionAES128, EncryptionAES192, EncryptionAES256, EncryptionSampleAES:
		return NewAESEncryptor()
	default:
		return nil
	}
}

// NewDecryptor creates a decryptor based on the method
func NewDecryptor(method EncryptionMethod) Decryptor {
	switch method {
	case EncryptionAES128, EncryptionAES192, EncryptionAES256, EncryptionSampleAES:
		return NewAESDecryptor()
	default:
		return nil
	}
}