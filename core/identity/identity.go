package identity

import (
	"encoding/base64"
	"fmt"

	"github.com/sureshKrishna05/aegisq-framework/core/aqx"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
)

type NodeIdentity struct {
	NodeID      string
	ValidatorID uint64
	PublicKey   []byte
	PrivateKey  []byte
	Signer      crypto.Signer
}

func NewNodeIdentity(nodeID string, validatorID uint64, signer crypto.Signer) (*NodeIdentity, error) {
	pub, priv, err := signer.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	return &NodeIdentity{
		NodeID:      nodeID,
		ValidatorID: validatorID,
		PublicKey:   pub,
		PrivateKey:  priv,
		Signer:      signer,
	}, nil
}

func (n *NodeIdentity) Sign(message []byte) ([]byte, error) {
	return n.Signer.Sign(n.PrivateKey, message)
}

func (n *NodeIdentity) Verify(message []byte, signature []byte) bool {
	return n.Signer.Verify(n.PublicKey, message, signature)
}

func (n *NodeIdentity) Algorithm() string {
	return n.Signer.Algorithm()
}

func (n *NodeIdentity) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(n.PublicKey)
}

func (n *NodeIdentity) String() string {
	return fmt.Sprintf(
		"NodeID: %s\nValidatorID: %d\nPublicKey: %s\nAlgorithm: %s\n",
		n.NodeID,
		n.ValidatorID,
		n.PublicKeyBase64(),
		n.Algorithm(),
	)
}

// =====================================================================
// AQX NETWORK LAYER (RFC type_id: 5 - Validator)
// =====================================================================

// ValidatorRecord represents the public-facing identity of a network node
// parsed from an AQX network broadcast. It explicitly lacks signing capabilities.
type ValidatorRecord struct {
	NodeID      string
	ValidatorID uint64
	PublicKey   []byte
}

// SerializePublicAQX creates the deterministic AQX representation of this node
// for broadcasting to the network as a "Validator" record.
// CRITICAL SECURITY: This NEVER serializes the PrivateKey.
func (n *NodeIdentity) SerializePublicAQX() []byte {
	e := aqx.AcquireEncoder()
	defer e.Release()

	// Strict AQX RFC Order
	e.String(n.NodeID)
	e.UInt64(n.ValidatorID)
	e.BytesArray(n.PublicKey)

	out := make([]byte, len(e.Bytes()))
	copy(out, e.Bytes())
	return out
}

// DeserializeValidatorAQX inflates raw AQX bytes back into a safe ValidatorRecord.
func DeserializeValidatorAQX(data []byte) (*ValidatorRecord, error) {
	d := aqx.NewDecoder(data)
	vr := &ValidatorRecord{}
	var err error

	if vr.NodeID, err = d.String(); err != nil {
		return nil, err
	}

	if vr.ValidatorID, err = d.UInt64(); err != nil {
		return nil, err
	}

	if vr.PublicKey, err = d.BytesArray(); err != nil {
		return nil, err
	}

	return vr, nil
}
