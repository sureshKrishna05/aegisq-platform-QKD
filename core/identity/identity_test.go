package identity

import (
	"bytes"
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
)

func TestIdentitySignVerify(t *testing.T) {
	signer := &crypto.Ed25519Signer{}

	node, err := NewNodeIdentity("validator-1", 1, signer)
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("Test message")

	signature, err := node.Sign(message)
	if err != nil {
		t.Fatal(err)
	}

	if !node.Verify(message, signature) {
		t.Fatal("Signature verification failed")
	}
}

func TestSignatureFailsOnModifiedMessage(t *testing.T) {
	signer := &crypto.Ed25519Signer{}

	node, _ := NewNodeIdentity("validator-1", 1, signer)

	message := []byte("Original message")
	signature, _ := node.Sign(message)

	modified := []byte("Tampered message")

	if node.Verify(modified, signature) {
		t.Fatal("Signature should fail for modified message")
	}
}

// =====================================================================
// AQX INTEGRATION TESTS
// =====================================================================

func TestSerializePublicAQX(t *testing.T) {
	signer := &crypto.Ed25519Signer{}

	node, err := NewNodeIdentity("validator-1", 1, signer)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Serialize the public identity using AQX
	rawBytes := node.SerializePublicAQX()

	// 2. Deserialize it back into a safe ValidatorRecord
	record, err := DeserializeValidatorAQX(rawBytes)
	if err != nil {
		t.Fatalf("Failed to deserialize AQX identity: %v", err)
	}

	// 3. Verify fields match exactly
	if record.NodeID != node.NodeID {
		t.Errorf("NodeID mismatch. Expected %s, got %s", node.NodeID, record.NodeID)
	}

	if !bytes.Equal(record.PublicKey, node.PublicKey) {
		t.Error("PublicKey mismatch after AQX deserialization")
	}

	// Structural Proof: `record` is a ValidatorRecord, which has no PrivateKey field.
	// This ensures the private key was perfectly stripped during serialization!
}
