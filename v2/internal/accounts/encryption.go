package accounts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"

	"tarkov-account-switcher/internal/config"
)

var (
	encryptionKey []byte
	keyOnce       sync.Once
	keyErr        error
)

// GetOrCreateKey loads or creates the encryption key exactly once.
func GetOrCreateKey() ([]byte, error) {
	keyOnce.Do(func() {
		paths := config.GetPaths()
		key, err := os.ReadFile(paths.KeyFile)
		if err == nil && len(key) == 32 {
			encryptionKey = key
			return
		}
		// Only create a key when none exists. Replacing a key that merely
		// couldn't be read (e.g. locked by a scanner) would make every saved
		// session undecryptable.
		if err == nil {
			keyErr = errors.New("encryption key file is damaged")
			return
		}
		if !os.IsNotExist(err) {
			keyErr = err
			return
		}
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			keyErr = err
			return
		}
		if err := os.WriteFile(paths.KeyFile, key, 0600); err != nil {
			keyErr = err
			return
		}
		encryptionKey = key
	})
	return encryptionKey, keyErr
}

// Encrypt encrypts plaintext using AES-256-CBC
// Returns format: iv_hex:encrypted_hex
func Encrypt(plaintext string) (string, error) {
	key, err := GetOrCreateKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	// Pad plaintext to block size (PKCS7)
	plainBytes := []byte(plaintext)
	padding := aes.BlockSize - (len(plainBytes) % aes.BlockSize)
	for range padding {
		plainBytes = append(plainBytes, byte(padding))
	}

	// Encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plainBytes))
	mode.CryptBlocks(encrypted, plainBytes)

	// Return iv:encrypted format
	return hex.EncodeToString(iv) + ":" + hex.EncodeToString(encrypted), nil
}

// Decrypt decrypts ciphertext that was encrypted with Encrypt
// Expects format: iv_hex:encrypted_hex
func Decrypt(ciphertext string) (string, error) {
	key, err := GetOrCreateKey()
	if err != nil {
		return "", err
	}

	parts := strings.Split(ciphertext, ":")
	if len(parts) != 2 {
		return "", errors.New("invalid ciphertext format")
	}

	iv, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", err
	}

	encrypted, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(encrypted)%aes.BlockSize != 0 {
		return "", errors.New("encrypted data is not a multiple of block size")
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Remove PKCS7 padding
	if len(decrypted) == 0 {
		return "", errors.New("decrypted data is empty")
	}
	padding := int(decrypted[len(decrypted)-1])
	if padding > aes.BlockSize || padding > len(decrypted) {
		return "", errors.New("invalid padding")
	}
	decrypted = decrypted[:len(decrypted)-padding]

	return string(decrypted), nil
}
