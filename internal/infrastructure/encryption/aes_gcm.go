package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Service defines the encryption interface.
type Service interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
	EncryptWithKey(plaintext string, key []byte) (string, error)
	DecryptWithKey(ciphertext string, key []byte) (string, error)
}

type aesGCMService struct {
	key []byte
}

// NewAESGCMService creates a new AES-256-GCM encryption service.
// encryptionKey must be at least 32 characters; only the first 32 bytes are used.
func NewAESGCMService(encryptionKey string) (Service, error) {
	if len(encryptionKey) < 32 {
		return nil, errors.New("encryption key must be at least 32 characters")
	}
	return &aesGCMService{key: []byte(encryptionKey)[:32]}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Output format: base64(nonce + ciphertext + auth_tag)
func (s *aesGCMService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM ciphertext.
func (s *aesGCMService) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.New("invalid ciphertext encoding")
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, data := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", errors.New("failed to decrypt secret")
	}

	return string(plaintext), nil
}

// EncryptWithKey encrypts plaintext using AES-256-GCM with the provided key.
// The key must be exactly 32 bytes. Output format: base64(nonce + ciphertext + auth_tag).
func (s *aesGCMService) EncryptWithKey(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be exactly 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithKey decrypts a base64-encoded AES-256-GCM ciphertext using the provided key.
// The key must be exactly 32 bytes.
func (s *aesGCMService) DecryptWithKey(ciphertext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be exactly 32 bytes")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.New("invalid ciphertext encoding")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, data := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", errors.New("failed to decrypt secret")
	}

	return string(plaintext), nil
}

