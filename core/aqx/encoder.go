package aqx

import (
	"encoding/binary"
	"sync"
)

const DefaultBufferSize = 1024 // 1KB covers most transactions and votes

// encoderPool completely eliminates heap allocations by recycling buffers
var encoderPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate the capacity to avoid runtime slice growth
		b := make([]byte, 0, DefaultBufferSize)
		return &Encoder{buf: b}
	},
}

// Encoder represents the AQX single-buffer architecture
type Encoder struct {
	buf []byte
}

// AcquireEncoder fetches a clean, pre-allocated encoder from the pool.
// Callers MUST defer Release() after extracting the Bytes().
func AcquireEncoder() *Encoder {
	e := encoderPool.Get().(*Encoder)
	e.buf = e.buf[:0] // Reset length to 0, but retain the underlying capacity
	return e
}

// Reset clears the encoder for immediate reuse without returning it to the pool
func (e *Encoder) Reset() {
	e.buf = e.buf[:0]
}

// Release returns the encoder to the pool to be recycled.
func (e *Encoder) Release() {
	// Guard against memory leaks from abnormal, massive payloads
	if cap(e.buf) <= 65536 {
		encoderPool.Put(e)
	}
}

// Bytes returns the canonical binary payload
func (e *Encoder) Bytes() []byte {
	return e.buf
}

// --- AQX Primitive Types ---
// Note: These methods do not return errors because we are writing to an
// in-memory slice that grows automatically. This drastically simplifies calling code!

func (e *Encoder) UInt8(v uint8) {
	e.buf = append(e.buf, v)
}

func (e *Encoder) UInt16(v uint16) {
	e.buf = binary.LittleEndian.AppendUint16(e.buf, v)
}

func (e *Encoder) UInt32(v uint32) {
	e.buf = binary.LittleEndian.AppendUint32(e.buf, v)
}

func (e *Encoder) UInt64(v uint64) {
	e.buf = binary.LittleEndian.AppendUint64(e.buf, v)
}

// Int32 and Int64 are cleanly cast to their unsigned counterparts for bitwise writing
func (e *Encoder) Int32(v int32) {
	e.buf = binary.LittleEndian.AppendUint32(e.buf, uint32(v))
}

func (e *Encoder) Int64(v int64) {
	e.buf = binary.LittleEndian.AppendUint64(e.buf, uint64(v))
}

func (e *Encoder) Bool(v bool) {
	if v {
		e.buf = append(e.buf, 1)
	} else {
		e.buf = append(e.buf, 0)
	}
}

// String implements the AQX length-prefixed text rule
func (e *Encoder) String(v string) {
	e.UInt32(uint32(len(v)))
	e.buf = append(e.buf, v...)
}

// Bytes implements the AQX length-prefixed raw bytes rule
func (e *Encoder) BytesArray(v []byte) {
	e.UInt32(uint32(len(v)))
	e.buf = append(e.buf, v...)
}

// FixedBytes writes raw bytes WITHOUT a length prefix.
// Used specifically for Hashes, Signatures, and Public Keys where the size is known.
func (e *Encoder) FixedBytes(v []byte) {
	e.buf = append(e.buf, v...)
}
