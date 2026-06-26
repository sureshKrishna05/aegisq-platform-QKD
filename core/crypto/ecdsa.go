package crypto

import (
	"crypto/sha256"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type ECDSASigner struct{}

func NewECDSASigner() (*ECDSASigner, error) {
	return &ECDSASigner{}, nil
}

func (e *ECDSASigner) GenerateKeyPair() ([]byte, []byte, error) {
	// Generate a new secp256k1 private key using go-ethereum
	privateKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	// 32-byte private key
	privBytes := ethcrypto.FromECDSA(privateKey)

	// 65-byte uncompressed public key
	pubBytes := ethcrypto.FromECDSAPub(&privateKey.PublicKey)

	return pubBytes, privBytes, nil
}

func (e *ECDSASigner) Sign(privateKey []byte, message []byte) ([]byte, error) {
	priv, err := ethcrypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	// Ethereum uses Keccak256, but we stick to SHA-256 for framework consistency 
	// (or we could use Keccak256 if preferred, but SHA-256 matches our existing Hash() functions).
	hash := sha256.Sum256(message)

	// Returns a 65-byte signature [R || S || V]
	sig, err := ethcrypto.Sign(hash[:], priv)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (e *ECDSASigner) Verify(publicKey, message, signature []byte) bool {
	// We need exactly the 64-byte (R || S) signature for VerifySignature
	var sigToVerify []byte
	if len(signature) == 65 {
		sigToVerify = signature[:64]
	} else if len(signature) == 64 {
		sigToVerify = signature
	} else {
		return false
	}

	hash := sha256.Sum256(message)

	// VerifySignature takes: pubkey (65 bytes), hash (32 bytes), sig (64 bytes)
	return ethcrypto.VerifySignature(publicKey, hash[:], sigToVerify)
}

func (e *ECDSASigner) Algorithm() string {
	return "ECDSA_secp256k1"
}
