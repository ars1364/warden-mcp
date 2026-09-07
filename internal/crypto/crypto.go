// Package crypto provides AES-256-GCM envelope encryption for secret values at rest.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Box struct {
	gcm cipher.AEAD
}

// NewBox builds an encryption box from a base64-encoded 32-byte master key.
func NewBox(base64Key string) (*Box, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("decode master key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("init cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("init gcm: %w", err)
	}

	return &Box{gcm: gcm}, nil
}

// Seal encrypts plaintext, returning nonce and ciphertext separately for storage.
func (b *Box) Seal(plaintext string) (nonce, ciphertext []byte, err error) {
	nonce = make([]byte, b.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext = b.gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return nonce, ciphertext, nil
}

// Open decrypts ciphertext using the given nonce.
func (b *Box) Open(nonce, ciphertext []byte) (string, error) {
	plaintext, err := b.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}

// GenerateMasterKey returns a new base64-encoded 32-byte key, for first-time setup.
func GenerateMasterKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
