package qkd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// SecureChannel represents an AES-256 encrypted tunnel between two validators
// established using a session key derived from QKD.
type SecureChannel struct {
	SessionKey []byte
	block      cipher.Block
}

func NewSecureChannelFromHex(hexKey string) (*SecureChannel, error) {
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	if len(keyBytes) != 32 {
		return nil, errors.New("AES key must be exactly 32 bytes (256-bit)")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	return &SecureChannel{
		SessionKey: keyBytes,
		block:      block,
	}, nil
}

// Encrypt payload using AES-256-GCM.
func (sc *SecureChannel) Encrypt(plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(sc.block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt payload using AES-256-GCM.
func (sc *SecureChannel) Decrypt(ciphertext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(sc.block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
