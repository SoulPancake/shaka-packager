// Copyright 2015 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"errors"
	"io"
	"sync"
	"time"
)

// IoCache implements a thread-safe circular buffer for I/O operations
type IoCache struct {
	mu             sync.RWMutex
	cacheSize      uint64
	circularBuffer []byte
	rPtr           uint64 // read pointer
	wPtr           uint64 // write pointer
	closed         bool
	closedForWrite bool
	
	// Condition variables using channels
	readSignal  chan struct{}
	writeSignal chan struct{}
}

// NewIoCache creates a new I/O cache with the specified size
func NewIoCache(cacheSize uint64) *IoCache {
	return &IoCache{
		cacheSize:      cacheSize,
		circularBuffer: make([]byte, cacheSize),
		readSignal:     make(chan struct{}, 1),
		writeSignal:    make(chan struct{}, 1),
	}
}

// Read reads data from the cache. This function may block until there is data in the cache.
func (ic *IoCache) Read(buffer []byte) (int64, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	
	for {
		ic.mu.Lock()
		if ic.closed && ic.BytesCachedInternal() == 0 {
			ic.mu.Unlock()
			return 0, io.EOF
		}
		
		available := ic.BytesCachedInternal()
		if available > 0 {
			// Read available data
			toRead := uint64(len(buffer))
			if toRead > available {
				toRead = available
			}
			
			// Handle circular buffer wrap-around
			readSize := toRead
			if ic.rPtr+toRead > ic.cacheSize {
				// Split read into two parts
				firstPart := ic.cacheSize - ic.rPtr
				copy(buffer[:firstPart], ic.circularBuffer[ic.rPtr:ic.cacheSize])
				copy(buffer[firstPart:toRead], ic.circularBuffer[0:toRead-firstPart])
				ic.rPtr = toRead - firstPart
			} else {
				copy(buffer[:toRead], ic.circularBuffer[ic.rPtr:ic.rPtr+toRead])
				ic.rPtr = (ic.rPtr + toRead) % ic.cacheSize
			}
			
			// Signal writers that space is available
			select {
			case ic.writeSignal <- struct{}{}:
			default:
			}
			
			ic.mu.Unlock()
			return int64(readSize), nil
		}
		
		ic.mu.Unlock()
		
		// Wait for data to become available
		select {
		case <-ic.readSignal:
			continue
		case <-time.After(100 * time.Millisecond):
			// Timeout to check closed status
			continue
		}
	}
}

// Write writes data to the cache. This function may block until there is enough room in the cache.
func (ic *IoCache) Write(buffer []byte) (int64, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	
	ic.mu.RLock()
	if ic.closedForWrite {
		ic.mu.RUnlock()
		return 0, errors.New("cache closed for writing")
	}
	ic.mu.RUnlock()
	
	totalWritten := int64(0)
	remaining := buffer
	
	for len(remaining) > 0 {
		ic.mu.Lock()
		if ic.closedForWrite {
			ic.mu.Unlock()
			return totalWritten, errors.New("cache closed for writing")
		}
		
		available := ic.BytesFreeInternal()
		if available > 0 {
			// Write available data
			toWrite := uint64(len(remaining))
			if toWrite > available {
				toWrite = available
			}
			
			// Handle circular buffer wrap-around
			if ic.wPtr+toWrite > ic.cacheSize {
				// Split write into two parts
				firstPart := ic.cacheSize - ic.wPtr
				copy(ic.circularBuffer[ic.wPtr:ic.cacheSize], remaining[:firstPart])
				copy(ic.circularBuffer[0:toWrite-firstPart], remaining[firstPart:toWrite])
				ic.wPtr = toWrite - firstPart
			} else {
				copy(ic.circularBuffer[ic.wPtr:ic.wPtr+toWrite], remaining[:toWrite])
				ic.wPtr = (ic.wPtr + toWrite) % ic.cacheSize
			}
			
			remaining = remaining[toWrite:]
			totalWritten += int64(toWrite)
			
			// Signal readers that data is available
			select {
			case ic.readSignal <- struct{}{}:
			default:
			}
		}
		
		ic.mu.Unlock()
		
		// If we still have data to write, wait for space
		if len(remaining) > 0 {
			select {
			case <-ic.writeSignal:
				continue
			case <-time.After(100 * time.Millisecond):
				// Timeout to check closed status
				continue
			}
		}
	}
	
	return totalWritten, nil
}

// Clear empties the cache
func (ic *IoCache) Clear() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.rPtr = 0
	ic.wPtr = 0
	
	// Signal any waiting operations
	select {
	case ic.writeSignal <- struct{}{}:
	default:
	}
}

// Close closes the cache
func (ic *IoCache) Close() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.closed = true
	ic.closedForWrite = true
	
	// Signal all waiting operations
	select {
	case ic.readSignal <- struct{}{}:
	default:
	}
	select {
	case ic.writeSignal <- struct{}{}:
	default:
	}
}

// CloseForWriting closes the cache for writing only
func (ic *IoCache) CloseForWriting() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.closedForWrite = true
	
	// Signal readers that no more data will be written
	select {
	case ic.readSignal <- struct{}{}:
	default:
	}
}

// Closed returns true if the cache is closed
func (ic *IoCache) Closed() bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	return ic.closed
}

// Reopen reopens the cache. Any data still in the cache will be lost.
func (ic *IoCache) Reopen() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.closed = false
	ic.closedForWrite = false
	ic.rPtr = 0
	ic.wPtr = 0
}

// BytesCached returns the number of bytes in the cache
func (ic *IoCache) BytesCached() uint64 {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	return ic.BytesCachedInternal()
}

// BytesAvailable is an alias for BytesCached for compatibility
func (ic *IoCache) BytesAvailable() uint64 {
	return ic.BytesCached()
}

// BytesFree returns the number of free bytes in the cache
func (ic *IoCache) BytesFree() uint64 {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	return ic.BytesFreeInternal()
}

// Flush flushes any pending operations
func (ic *IoCache) Flush() error {
	// Signal any waiting operations
	select {
	case ic.readSignal <- struct{}{}:
	default:
	}
	select {
	case ic.writeSignal <- struct{}{}:
	default:
	}
	
	return nil
}

// WaitUntilEmptyOrClosed waits until the cache is empty or has been closed
func (ic *IoCache) WaitUntilEmptyOrClosed() {
	for {
		ic.mu.RLock()
		if ic.closed || ic.BytesCachedInternal() == 0 {
			ic.mu.RUnlock()
			return
		}
		ic.mu.RUnlock()
		
		time.Sleep(1 * time.Millisecond)
	}
}

// BytesCachedInternal returns the number of cached bytes (must be called with lock held)
func (ic *IoCache) BytesCachedInternal() uint64 {
	if ic.wPtr >= ic.rPtr {
		return ic.wPtr - ic.rPtr
	}
	return ic.cacheSize - ic.rPtr + ic.wPtr
}

// BytesFreeInternal returns the number of free bytes (must be called with lock held)  
func (ic *IoCache) BytesFreeInternal() uint64 {
	return ic.cacheSize - ic.BytesCachedInternal() - 1 // -1 to avoid full=empty ambiguity
}