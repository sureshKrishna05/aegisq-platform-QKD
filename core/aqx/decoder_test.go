package aqx

import (
	"bytes"
	"testing"
)

func TestDecoder_Primitives(t *testing.T) {
	e := AcquireEncoder()
	defer e.Release()

	// 1. Encode all primitive types
	e.UInt8(255)
	e.UInt16(65535)
	e.UInt32(4294967295)
	e.UInt64(18446744073709551615)
	e.Int32(-12345)
	e.Int64(-9876543210)
	e.Bool(true)
	e.Bool(false)

	// 2. Decode them in the exact same order
	d := NewDecoder(e.Bytes())

	if v, err := d.UInt8(); err != nil || v != 255 {
		t.Errorf("UInt8 failed: expected 255, got %v (err: %v)", v, err)
	}
	if v, err := d.UInt16(); err != nil || v != 65535 {
		t.Errorf("UInt16 failed: expected 65535, got %v (err: %v)", v, err)
	}
	if v, err := d.UInt32(); err != nil || v != 4294967295 {
		t.Errorf("UInt32 failed: expected 4294967295, got %v (err: %v)", v, err)
	}
	if v, err := d.UInt64(); err != nil || v != 18446744073709551615 {
		t.Errorf("UInt64 failed: expected 18446744073709551615, got %v (err: %v)", v, err)
	}
	if v, err := d.Int32(); err != nil || v != -12345 {
		t.Errorf("Int32 failed: expected -12345, got %v (err: %v)", v, err)
	}
	if v, err := d.Int64(); err != nil || v != -9876543210 {
		t.Errorf("Int64 failed: expected -9876543210, got %v (err: %v)", v, err)
	}
	if v, err := d.Bool(); err != nil || v != true {
		t.Errorf("Bool(true) failed: expected true, got %v (err: %v)", v, err)
	}
	if v, err := d.Bool(); err != nil || v != false {
		t.Errorf("Bool(false) failed: expected false, got %v (err: %v)", v, err)
	}
}

func TestDecoder_StringAndBytes(t *testing.T) {
	e := AcquireEncoder()
	defer e.Release()

	testStr := "AQX Protocol"
	testBytes := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	e.String(testStr)
	e.BytesArray(testBytes)

	d := NewDecoder(e.Bytes())

	str, err := d.String()
	if err != nil || str != testStr {
		t.Errorf("String failed: expected %s, got %s (err: %v)", testStr, str, err)
	}

	b, err := d.BytesArray()
	if err != nil || !bytes.Equal(b, testBytes) {
		t.Errorf("BytesArray failed: expected %x, got %x (err: %v)", testBytes, b, err)
	}

	// PROVE ZERO-COPY: Modifying the decoded slice should alter the original buffer
	b[0] = 0x00

	// Offset calculation: Length prefix of string (4) + string content (12) + Length prefix of bytes (4)
	originalBufferOffset := 4 + 12 + 4
	if e.Bytes()[originalBufferOffset] != 0x00 {
		t.Errorf("BytesArray is NOT zero-copy! Modifying the decoded slice did not modify the underlying encoder buffer.")
	}
}

func TestDecoder_FixedBytes(t *testing.T) {
	e := AcquireEncoder()
	defer e.Release()

	testFixed := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	e.FixedBytes(testFixed)

	d := NewDecoder(e.Bytes())

	b, err := d.FixedBytes(5) // Read exactly 5 bytes
	if err != nil || !bytes.Equal(b, testFixed) {
		t.Errorf("FixedBytes failed: expected %x, got %x (err: %v)", testFixed, b, err)
	}
}

func TestDecoder_OutOfBoundsSafety(t *testing.T) {
	// 1. Buffer too small for primitive
	d1 := NewDecoder([]byte{0x00, 0x01})
	_, err := d1.UInt32() // Tries to read 4 bytes
	if err != ErrOutOfBounds {
		t.Errorf("Expected ErrOutOfBounds for UInt32, got %v", err)
	}

	// 2. Malicious length prefix
	// The prefix (0x0A000000 -> 10 bytes) claims the string is 10 bytes long,
	// but the buffer ends immediately after the prefix.
	d2 := NewDecoder([]byte{0x0A, 0x00, 0x00, 0x00, 0xFF})
	_, err = d2.String()
	if err != ErrOutOfBounds {
		t.Errorf("Expected ErrOutOfBounds for malicious string length, got %v", err)
	}

	// 3. FixedBytes boundary check
	d3 := NewDecoder([]byte{0x01, 0x02})
	_, err = d3.FixedBytes(5)
	if err != ErrOutOfBounds {
		t.Errorf("Expected ErrOutOfBounds for FixedBytes, got %v", err)
	}
}
