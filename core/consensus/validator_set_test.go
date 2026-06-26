package consensus

import "testing"

func TestAddAndAuthorizeValidator(t *testing.T) {
	vs := NewValidatorSet()
	pub := []byte("validator_public_key")

	vs.AddValidator(1, pub)

	if !vs.IsAuthorized(1, pub) {
		t.Fatal("validator should be authorized")
	}
}

func TestUnauthorizedValidator(t *testing.T) {
	vs := NewValidatorSet()
	pub := []byte("validator_public_key")

	vs.AddValidator(1, pub)

	if vs.IsAuthorized(2, pub) {
		t.Fatal("unknown validator should not be authorized")
	}
}

func TestWrongPublicKeyRejected(t *testing.T) {
	vs := NewValidatorSet()
	pub := []byte("correct_key")
	wrong := []byte("wrong_key")

	vs.AddValidator(1, pub)

	if vs.IsAuthorized(1, wrong) {
		t.Fatal("authorization should fail for wrong public key")
	}
}

func TestRemoveValidator(t *testing.T) {
	vs := NewValidatorSet()
	pub := []byte("validator_public_key")

	vs.AddValidator(1, pub)
	vs.RemoveValidator(1)

	if vs.IsAuthorized(1, pub) {
		t.Fatal("removed validator should not be authorized")
	}
}

func TestValidatorCount(t *testing.T) {
	vs := NewValidatorSet()

	vs.AddValidator(1, []byte("key1"))
	vs.AddValidator(2, []byte("key2"))

	if vs.Count() != 2 {
		t.Fatal("validator count incorrect")
	}
}
