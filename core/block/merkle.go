package block

import (
	"github.com/sureshKrishna05/aegisq-framework/core/aqx"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
)

// ComputeMerkleRoot recursively calculates the root hash of a list of transaction hashes.
// Optimized using the AQX zero-allocation memory pool.
func ComputeMerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return nil
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// OPTIMIZATION 1: Pre-allocate the nextLevel slice capacity.
	// Since we are pairing hashes, the next level will be exactly half the size
	// (plus one if the current length is odd). This prevents `append` from re-allocating memory!
	nextLevel := make([][]byte, 0, (len(hashes)+1)/2)

	for i := 0; i < len(hashes); i += 2 {
		if i+1 == len(hashes) {
			// Odd number of hashes, carry the last one over to the next level
			nextLevel = append(nextLevel, hashes[i])
		} else {
			// OPTIMIZATION 2: Use the AQX memory pool instead of bytes.Join()
			e := aqx.AcquireEncoder()

			e.FixedBytes(hashes[i])
			e.FixedBytes(hashes[i+1])

			// Hash the concatenated bytes
			nextLevel = append(nextLevel, crypto.Hash(e.Bytes()))

			// Release the encoder IMMEDIATELY so it can be reused in the very next loop iteration!
			e.Release()
		}
	}

	return ComputeMerkleRoot(nextLevel)
}
