package aqx

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncoder_Primitives(t *testing.T) {
	e := AcquireEncoder()
	defer e.Release()

	// Write basic primitives
	e.UInt8(255)                   // 1 byte: 0xFF
	e.UInt16(65535)                // 2 bytes: 0xFF 0xFF
	e.UInt32(4294967295)           // 4 bytes: 0xFF 0xFF 0xFF 0xFF
	e.UInt64(18446744073709551615) // 8 bytes: 0xFF x 8
	e.Bool(true)                   // 1 byte: 0x01
	e.Bool(false)                  // 1 byte: 0x00

	expectedLen := 1 + 2 + 4 + 8 + 1 + 1
	res := e.Bytes()

	if len(res) != expectedLen {
		t.Fatalf("Expected length %d, got %d", expectedLen, len(res))
	}

	// Verify little-endian packing of the UInt16
	if res[1] != 255 || res[2] != 255 {
		t.Errorf("UInt16 packing failed")
	}

	// Verify boolean packing at the end
	if res[15] != 1 {
		t.Errorf("Expected true to be 1")
	}
	if res[16] != 0 {
		t.Errorf("Expected false to be 0")
	}
}

func TestEncoder_StringAndBytes(t *testing.T) {
	e := AcquireEncoder()
	defer e.Release()

	testStr := "AQX"
	e.String(testStr)

	res := e.Bytes()

	// Expected: 4 bytes for length (uint32) + 3 bytes for "AQX"
	if len(res) != 7 {
		t.Fatalf("Expected length 7, got %d", len(res))
	}

	// Extract the length prefix
	strLen := binary.LittleEndian.Uint32(res[0:4])
	if strLen != 3 {
		t.Errorf("Expected length prefix 3, got %d", strLen)
	}

	// Extract the string contents
	if string(res[4:7]) != "AQX" {
		t.Errorf("Expected AQX, got %s", string(res[4:7]))
	}
}

func TestEncoder_FixedBytes(t *testing.T) {
	e := AcquireEncoder()
	defer e.Release()

	hash := []byte{0x01, 0x02, 0x03, 0x04}
	e.FixedBytes(hash)

	res := e.Bytes()

	// Fixed bytes should NOT have a length prefix
	if len(res) != 4 {
		t.Fatalf("Expected length 4, got %d", len(res))
	}
	if !bytes.Equal(res, hash) {
		t.Errorf("FixedBytes corrupted the input data")
	}
}

func TestEncoder_PoolRecycling(t *testing.T) {
	// 1. Acquire an encoder and write some data
	e1 := AcquireEncoder()
	e1.String("Heavy Payload Data")
	originalCap := cap(e1.Bytes())

	// 2. Release it back to the pool
	e1.Release()

	// 3. Acquire a new encoder (which should pull the exact same one from the pool)
	e2 := AcquireEncoder()
	defer e2.Release()

	// Ensure the length was reset to 0 so we don't accidentally read old data
	if len(e2.Bytes()) != 0 {
		t.Fatalf("Pooled encoder was not reset! Length is %d", len(e2.Bytes()))
	}

	// Ensure the capacity was retained so we don't trigger GC allocations
	if cap(e2.Bytes()) != originalCap {
		t.Fatalf("Pooled encoder lost its capacity. Expected %d, got %d", originalCap, cap(e2.Bytes()))
	}
}
