package aqx

import (
	"encoding/binary"
	"errors"
)

// ErrOutOfBounds prevents malicious payloads from causing a node to panic
var ErrOutOfBounds = errors.New("aqx: read out of bounds")

// Decoder reads AQX canonical binary payloads using a fast, zero-copy cursor
type Decoder struct {
	buf    []byte
	offset int
}

// NewDecoder initializes a cursor for reading
func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		buf:    data,
		offset: 0,
	}
}

// --- AQX Primitive Decoders ---
// Notice how every function strictly checks the bounds before reading.
// This prevents network-level crash attacks.

func (d *Decoder) UInt8() (uint8, error) {
	if d.offset+1 > len(d.buf) {
		return 0, ErrOutOfBounds
	}
	v := d.buf[d.offset]
	d.offset++
	return v, nil
}

func (d *Decoder) UInt16() (uint16, error) {
	if d.offset+2 > len(d.buf) {
		return 0, ErrOutOfBounds
	}
	v := binary.LittleEndian.Uint16(d.buf[d.offset:])
	d.offset += 2
	return v, nil
}

func (d *Decoder) UInt32() (uint32, error) {
	if d.offset+4 > len(d.buf) {
		return 0, ErrOutOfBounds
	}
	v := binary.LittleEndian.Uint32(d.buf[d.offset:])
	d.offset += 4
	return v, nil
}

func (d *Decoder) UInt64() (uint64, error) {
	if d.offset+8 > len(d.buf) {
		return 0, ErrOutOfBounds
	}
	v := binary.LittleEndian.Uint64(d.buf[d.offset:])
	d.offset += 8
	return v, nil
}

func (d *Decoder) Int32() (int32, error) {
	v, err := d.UInt32()
	return int32(v), err
}

func (d *Decoder) Int64() (int64, error) {
	v, err := d.UInt64()
	return int64(v), err
}

func (d *Decoder) Bool() (bool, error) {
	v, err := d.UInt8()
	return v == 1, err
}

// String extracts the length prefix first, then reads the text
func (d *Decoder) String() (string, error) {
	length, err := d.UInt32()
	if err != nil {
		return "", err
	}

	if d.offset+int(length) > len(d.buf) {
		return "", ErrOutOfBounds
	}

	v := string(d.buf[d.offset : d.offset+int(length)])
	d.offset += int(length)
	return v, nil
}

// BytesArray implements the ZERO-COPY rule!
// It returns a slice pointing directly to the original buffer.
func (d *Decoder) BytesArray() ([]byte, error) {
	length, err := d.UInt32()
	if err != nil {
		return nil, err
	}

	if d.offset+int(length) > len(d.buf) {
		return nil, ErrOutOfBounds
	}

	v := d.buf[d.offset : d.offset+int(length)]
	d.offset += int(length)
	return v, nil
}

// FixedBytes reads a known length WITHOUT looking for a length prefix.
// ZERO-COPY rule applies here too.
func (d *Decoder) FixedBytes(length int) ([]byte, error) {
	if d.offset+length > len(d.buf) {
		return nil, ErrOutOfBounds
	}

	v := d.buf[d.offset : d.offset+length]
	d.offset += length
	return v, nil
}
