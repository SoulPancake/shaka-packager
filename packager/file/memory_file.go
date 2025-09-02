// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"bytes"
	"io"
	"strings"
	"sync"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

// MemoryFile represents an in-memory file.
type MemoryFile struct {
	*BaseFile
	buffer   *bytes.Buffer
	reader   *bytes.Reader
	mode     string
	position uint64
	mu       sync.RWMutex
	closed   bool
}

// NewMemoryFile creates a new memory file.
func NewMemoryFile(fileName, mode string) (*MemoryFile, packager.Status) {
	// Strip memory:// prefix if present
	actualFileName := fileName
	if strings.HasPrefix(fileName, MemoryFilePrefix) {
		actualFileName = fileName[len(MemoryFilePrefix):]
	}

	mf := &MemoryFile{
		BaseFile: NewBaseFile(actualFileName),
		buffer:   new(bytes.Buffer),
		mode:     mode,
	}

	return mf, packager.StatusOK
}

// Close closes the memory file.
func (mf *MemoryFile) Close() bool {
	mf.mu.Lock()
	defer mf.mu.Unlock()
	
	mf.closed = true
	return true
}

// Read reads data from the memory file.
func (mf *MemoryFile) Read(buffer []byte) (int64, error) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	if mf.closed {
		return 0, packager.NewStatus(packager.FILE_FAILURE, "File closed").Error()
	}

	// For read operations, create a reader if needed
	if mf.reader == nil {
		mf.reader = bytes.NewReader(mf.buffer.Bytes())
	}

	n, err := mf.reader.Read(buffer)
	mf.position += uint64(n)

	return int64(n), err
}

// Write writes data to the memory file.
func (mf *MemoryFile) Write(buffer []byte) (int64, error) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	if mf.closed {
		return 0, packager.NewStatus(packager.FILE_FAILURE, "File closed").Error()
	}

	n, err := mf.buffer.Write(buffer)
	mf.position += uint64(n)

	// Reset reader since buffer changed
	mf.reader = nil

	return int64(n), err
}

// CloseForWriting closes the file for writing.
func (mf *MemoryFile) CloseForWriting() {
	// For memory files, this is a no-op
}

// Size returns the size of the memory file.
func (mf *MemoryFile) Size() int64 {
	mf.mu.RLock()
	defer mf.mu.RUnlock()

	return int64(mf.buffer.Len())
}

// Flush flushes the memory file (no-op for memory files).
func (mf *MemoryFile) Flush() bool {
	return true
}

// Seek seeks to a position in the memory file.
func (mf *MemoryFile) Seek(position uint64) bool {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	if mf.closed {
		return false
	}

	// Create reader if needed
	if mf.reader == nil {
		mf.reader = bytes.NewReader(mf.buffer.Bytes())
	}

	_, err := mf.reader.Seek(int64(position), io.SeekStart)
	if err != nil {
		return false
	}

	mf.position = position
	return true
}

// Tell returns the current position in the memory file.
func (mf *MemoryFile) Tell() (uint64, bool) {
	mf.mu.RLock()
	defer mf.mu.RUnlock()

	return mf.position, !mf.closed
}

// GetBytes returns the underlying byte slice (for testing/debugging).
func (mf *MemoryFile) GetBytes() []byte {
	mf.mu.RLock()
	defer mf.mu.RUnlock()

	return mf.buffer.Bytes()
}