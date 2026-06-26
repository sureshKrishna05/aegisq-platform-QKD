package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
)

type Ed25519Signer struct{}

func (e *Ed25519Signer) GenerateKeyPair() ([]byte, []byte, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	return pub, priv, err
}

func (e *Ed25519Signer) Sign(privateKey []byte, message []byte) ([]byte, error) {
	priv := ed25519.PrivateKey(privateKey)

	// Note: Ed25519 hashes the message internally (with SHA-512)
	// as part of the protocol, so we pass the raw message directly.
	signature := ed25519.Sign(priv, message)

	return signature, nil
}

func (e *Ed25519Signer) Verify(publicKey []byte, message []byte, signature []byte) bool {
	// Strict rejection of invalid signature lengths (Ed25519 is always 64 bytes)
	if len(signature) != ed25519.SignatureSize {
		return false
	}

	pub := ed25519.PublicKey(publicKey)
	return ed25519.Verify(pub, message, signature)
}

func (e *Ed25519Signer) Algorithm() string {
	return "ed25519"
}
