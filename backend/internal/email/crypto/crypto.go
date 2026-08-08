// Package emailcrypto provides AES-256-GCM encryption for storing email credentials.
// The key is derived from the application's JWT_ACCESS_SECRET so no extra config is needed.
package emailcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// deriveKey produces a 32-byte AES-256 key from an arbitrary string secret.
func deriveKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a base64-encoded
// string of nonce+ciphertext.
func Encrypt(secret, plaintext string) (string, error) {
	key := deriveKey(secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create GCM: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("emailcrypto: generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded nonce+ciphertext produced by Encrypt.
func Decrypt(secret, encoded string) (string, error) {
	key := deriveKey(secret)

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: base64 decode: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create GCM: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("emailcrypto: ciphertext too short")
	}

	nonce, ct := data[:nonceSize], data[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: decrypt: %w", err)
	}

	return string(plaintext), nil
}
