package simulation

import (
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
)

// =====================================================================
// END-TO-END PROTOCOL THROUGHPUT BENCHMARKS
// =====================================================================

func BenchmarkE2E_TransactionLifecycle(b *testing.B) {
	signer, err := crypto.NewDilithiumSigner()
	if err != nil {
		b.Fatal(err)
	}
	defer signer.Close()

	node, err := identity.NewNodeIdentity("validator-1", 1, signer)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 1. Creation
		tx := transaction.NewTransaction(node, [32]byte{}, "PayloadData")
		tx.Nonce = uint64(i + 1)
		
		// 2. Hash & Sign
		if err := tx.SignWithIdentity(node); err != nil {
			b.Fatal(err)
		}

		// 3. Network Encode (AQX)
		rawBytes := tx.SerializeAQX()

		// 4. Network Decode (AQX)
		decodedTx, err := transaction.DeserializeAQX(rawBytes)
		if err != nil {
			b.Fatal(err)
		}

		// 5. Verify
		valid, err := decodedTx.Verify(signer)
		if err != nil || !valid {
			b.Fatal("Verification failed in E2E benchmark")
		}
	}
}
