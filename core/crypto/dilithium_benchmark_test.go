package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// =====================================================================
// PURE ALGORITHM BENCHMARKS (No Wrapper Overhead)
// =====================================================================

// --- DILITHIUM ---

func BenchmarkAlgo_DilithiumKeyGen(b *testing.B) {
	signer, _ := NewDilithiumSigner()
	defer signer.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.GenerateKeyPair()
	}
}

func BenchmarkAlgo_DilithiumSign(b *testing.B) {
	signer, _ := NewDilithiumSigner()
	defer signer.Close()
	_, priv, _ := signer.GenerateKeyPair()
	msg := []byte("Benchmark Message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Dilithium signer doesn't have costly private key parsing in its wrapper
		signer.Sign(priv, msg)
	}
}

func BenchmarkAlgo_DilithiumVerify(b *testing.B) {
	signer, _ := NewDilithiumSigner()
	defer signer.Close()
	pub, priv, _ := signer.GenerateKeyPair()
	msg := []byte("Benchmark Message")
	sig, _ := signer.Sign(priv, msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.Verify(pub, msg, sig)
	}
}

// --- ECDSA (secp256k1) ---

func BenchmarkAlgo_ECDSAKeyGen(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ethcrypto.GenerateKey()
	}
}

func BenchmarkAlgo_ECDSASign(b *testing.B) {
	priv, _ := ethcrypto.GenerateKey()
	msg := []byte("Benchmark Message")
	hash := sha256.Sum256(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Pure signing, no ToECDSA parsing
		ethcrypto.Sign(hash[:], priv)
	}
}

func BenchmarkAlgo_ECDSAVerify(b *testing.B) {
	priv, _ := ethcrypto.GenerateKey()
	pub := ethcrypto.FromECDSAPub(&priv.PublicKey)
	msg := []byte("Benchmark Message")
	hash := sha256.Sum256(msg)
	sig, _ := ethcrypto.Sign(hash[:], priv)
	sigToVerify := sig[:64]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ethcrypto.VerifySignature(pub, hash[:], sigToVerify)
	}
}

// --- ED25519 ---

func BenchmarkAlgo_Ed25519KeyGen(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.GenerateKey(rand.Reader)
	}
}

func BenchmarkAlgo_Ed25519Sign(b *testing.B) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	msg := []byte("Benchmark Message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(priv, msg)
	}
}

func BenchmarkAlgo_Ed25519Verify(b *testing.B) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	msg := []byte("Benchmark Message")
	sig := ed25519.Sign(priv, msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Verify(pub, msg, sig)
	}
}
