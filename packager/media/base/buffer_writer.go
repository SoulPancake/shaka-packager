// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package base

import (
	"bytes"
	"encoding/binary"

	"github.com/SoulPancake/shaka-packager/packager/status"
)

// File interface for writing
type File interface {
	Write([]byte) (int64, error)
}

// BufferWriter is a simple buffer writer that appends various data types to buffer
type BufferWriter struct {
	buf bytes.Buffer
}

// NewBufferWriter constructs a new BufferWriter
func NewBufferWriter() *BufferWriter {
	return &BufferWriter{}
}

// NewBufferWriterWithCapacity constructs a BufferWriter with reserved capacity
// The reservedSize is for optimization and doesn't affect the actual size
func NewBufferWriterWithCapacity(reservedSize int) *BufferWriter {
	writer := &BufferWriter{}
	writer.buf.Grow(reservedSize)
	return writer
}

// AppendUint8 appends a uint8 to the buffer
func (bw *BufferWriter) AppendUint8(v uint8) {
	bw.buf.WriteByte(v)
}

// AppendUint16 appends a uint16 in big-endian byte order
func (bw *BufferWriter) AppendUint16(v uint16) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], v)
	bw.buf.Write(buf[:])
}

// AppendUint32 appends a uint32 in big-endian byte order
func (bw *BufferWriter) AppendUint32(v uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	bw.buf.Write(buf[:])
}

// AppendUint64 appends a uint64 in big-endian byte order
func (bw *BufferWriter) AppendUint64(v uint64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	bw.buf.Write(buf[:])
}

// AppendInt16 appends an int16 in big-endian byte order
func (bw *BufferWriter) AppendInt16(v int16) {
	bw.AppendUint16(uint16(v))
}

// AppendInt32 appends an int32 in big-endian byte order
func (bw *BufferWriter) AppendInt32(v int32) {
	bw.AppendUint32(uint32(v))
}

// AppendInt64 appends an int64 in big-endian byte order
func (bw *BufferWriter) AppendInt64(v int64) {
	bw.AppendUint64(uint64(v))
}

// AppendNBytes appends the least significant numBytes of v to buffer
// numBytes should not be larger than 8
func (bw *BufferWriter) AppendNBytes(v uint64, numBytes int) {
	if numBytes <= 0 || numBytes > 8 {
		return
	}
	
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	bw.buf.Write(buf[8-numBytes:])
}

// AppendBytes appends a byte slice to the buffer
func (bw *BufferWriter) AppendBytes(v []byte) {
	bw.buf.Write(v)
}

// AppendString appends a string to the buffer
func (bw *BufferWriter) AppendString(s string) {
	bw.buf.WriteString(s)
}

// AppendArray appends an array to the buffer
func (bw *BufferWriter) AppendArray(data []byte) {
	bw.buf.Write(data)
}

// AppendBuffer appends another BufferWriter's contents
func (bw *BufferWriter) AppendBuffer(other *BufferWriter) {
	bw.buf.Write(other.Bytes())
}

// Swap swaps the internal buffer with another BufferWriter
func (bw *BufferWriter) Swap(other *BufferWriter) {
	// Since bytes.Buffer doesn't have a swap method, we copy the data
	thisBuf := bw.buf.Bytes()
	otherBuf := other.buf.Bytes()
	
	bw.buf.Reset()
	other.buf.Reset()
	
	bw.buf.Write(otherBuf)
	other.buf.Write(thisBuf)
}

// SwapBuffer swaps the internal buffer with a byte slice
func (bw *BufferWriter) SwapBuffer(buffer *[]byte) {
	currentBuf := bw.buf.Bytes()
	bw.buf.Reset()
	bw.buf.Write(*buffer)
	*buffer = currentBuf
}

// Clear clears the buffer
func (bw *BufferWriter) Clear() {
	bw.buf.Reset()
}

// Size returns the size of the buffer
func (bw *BufferWriter) Size() int {
	return bw.buf.Len()
}

// Bytes returns the underlying buffer data
func (bw *BufferWriter) Bytes() []byte {
	return bw.buf.Bytes()
}

// Buffer returns the underlying buffer data (alias for Bytes)
func (bw *BufferWriter) Buffer() []byte {
	return bw.buf.Bytes()
}

// WriteToFile writes the buffer to a file and clears the internal buffer
func (bw *BufferWriter) WriteToFile(file File) *status.Status {
	if file == nil {
		return status.NewStatus(status.InvalidArgument, "file cannot be nil")
	}
	
	data := bw.buf.Bytes()
	if len(data) == 0 {
		return status.NewOkStatus()
	}
	
	bytesWritten, err := file.Write(data)
	if err != nil {
		return status.NewStatus(status.Unknown, err.Error())
	}
	
	if bytesWritten != int64(len(data)) {
		return status.NewStatus(status.Unknown, "incomplete write")
	}
	
	bw.buf.Reset()
	return status.NewOkStatus()
}

// String returns the buffer contents as a string (for debugging)
func (bw *BufferWriter) String() string {
	return bw.buf.String()
}