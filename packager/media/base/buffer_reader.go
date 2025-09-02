// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package base

import (
	"encoding/binary"
)

// BufferReader reads data of various types from a fixed byte array
type BufferReader struct {
	buf  []byte
	size int
	pos  int
}

// NewBufferReader creates a BufferReader from a raw buffer
func NewBufferReader(buf []byte) *BufferReader {
	return &BufferReader{
		buf:  buf,
		size: len(buf),
		pos:  0,
	}
}

// HasBytes returns true if there are more than count bytes remaining
func (br *BufferReader) HasBytes(count int) bool {
	return br.pos+count <= br.size
}

// Read1 reads a uint8 value and advances the position
func (br *BufferReader) Read1() (uint8, bool) {
	if !br.HasBytes(1) {
		return 0, false
	}
	value := br.buf[br.pos]
	br.pos++
	return value, true
}

// Read2 reads a uint16 value in big-endian byte order
func (br *BufferReader) Read2() (uint16, bool) {
	if !br.HasBytes(2) {
		return 0, false
	}
	value := binary.BigEndian.Uint16(br.buf[br.pos:])
	br.pos += 2
	return value, true
}

// Read2s reads an int16 value in big-endian byte order
func (br *BufferReader) Read2s() (int16, bool) {
	value, ok := br.Read2()
	return int16(value), ok
}

// Read4 reads a uint32 value in big-endian byte order
func (br *BufferReader) Read4() (uint32, bool) {
	if !br.HasBytes(4) {
		return 0, false
	}
	value := binary.BigEndian.Uint32(br.buf[br.pos:])
	br.pos += 4
	return value, true
}

// Read4s reads an int32 value in big-endian byte order
func (br *BufferReader) Read4s() (int32, bool) {
	value, ok := br.Read4()
	return int32(value), ok
}

// Read8 reads a uint64 value in big-endian byte order
func (br *BufferReader) Read8() (uint64, bool) {
	if !br.HasBytes(8) {
		return 0, false
	}
	value := binary.BigEndian.Uint64(br.buf[br.pos:])
	br.pos += 8
	return value, true
}

// Read8s reads an int64 value in big-endian byte order
func (br *BufferReader) Read8s() (int64, bool) {
	value, ok := br.Read8()
	return int64(value), ok
}

// ReadNBytesInto8 reads N bytes as uint64 (big-endian)
func (br *BufferReader) ReadNBytesInto8(numBytes int) (uint64, bool) {
	if numBytes <= 0 || numBytes > 8 || !br.HasBytes(numBytes) {
		return 0, false
	}
	
	var value uint64
	for i := 0; i < numBytes; i++ {
		value = (value << 8) | uint64(br.buf[br.pos+i])
	}
	br.pos += numBytes
	return value, true
}

// ReadNBytesInto8s reads N bytes as int64 (big-endian)
func (br *BufferReader) ReadNBytesInto8s(numBytes int) (int64, bool) {
	value, ok := br.ReadNBytesInto8(numBytes)
	
	// Handle sign extension for negative values
	if ok && numBytes < 8 {
		// Check if the most significant bit is set
		signBit := uint64(1) << uint64(numBytes*8-1)
		if value&signBit != 0 {
			// Extend the sign
			mask := uint64(0xFFFFFFFFFFFFFFFF) << uint64(numBytes*8)
			value |= mask
		}
	}
	
	return int64(value), ok
}

// ReadToVector reads count bytes into a byte slice
func (br *BufferReader) ReadToVector(count int) ([]byte, bool) {
	if !br.HasBytes(count) {
		return nil, false
	}
	
	result := make([]byte, count)
	copy(result, br.buf[br.pos:br.pos+count])
	br.pos += count
	return result, true
}

// ReadToString reads size bytes into a string
func (br *BufferReader) ReadToString(size int) (string, bool) {
	if !br.HasBytes(size) {
		return "", false
	}
	
	result := string(br.buf[br.pos : br.pos+size])
	br.pos += size
	return result, true
}

// ReadCString reads a null-terminated string
func (br *BufferReader) ReadCString() (string, bool) {
	start := br.pos
	
	// Find the null terminator
	for i := start; i < br.size; i++ {
		if br.buf[i] == 0 {
			result := string(br.buf[start:i])
			br.pos = i + 1 // Skip the null terminator
			return result, true
		}
	}
	
	// No null terminator found
	return "", false
}

// SkipBytes advances the stream by numBytes
func (br *BufferReader) SkipBytes(numBytes int) bool {
	if !br.HasBytes(numBytes) {
		return false
	}
	br.pos += numBytes
	return true
}

// Data returns the underlying buffer
func (br *BufferReader) Data() []byte {
	return br.buf
}

// Size returns the buffer size
func (br *BufferReader) Size() int {
	return br.size
}

// SetSize sets the buffer size
func (br *BufferReader) SetSize(size int) {
	if size >= 0 && size <= len(br.buf) {
		br.size = size
	}
}

// Pos returns the current position
func (br *BufferReader) Pos() int {
	return br.pos
}

// SetPos sets the current position
func (br *BufferReader) SetPos(pos int) bool {
	if pos >= 0 && pos <= br.size {
		br.pos = pos
		return true
	}
	return false
}

// Remaining returns the number of bytes remaining
func (br *BufferReader) Remaining() int {
	return br.size - br.pos
}

// Reset resets the reader to the beginning
func (br *BufferReader) Reset() {
	br.pos = 0
}

// PeekBytes returns bytes at the current position without advancing
func (br *BufferReader) PeekBytes(count int) ([]byte, bool) {
	if !br.HasBytes(count) {
		return nil, false
	}
	return br.buf[br.pos : br.pos+count], true
}

// Peek1 peeks at a uint8 without advancing position
func (br *BufferReader) Peek1() (uint8, bool) {
	if !br.HasBytes(1) {
		return 0, false
	}
	return br.buf[br.pos], true
}