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
	"sync/atomic"
	"time"
)

// ThreadedIoMode represents the mode of threaded I/O operation
type ThreadedIoMode int

const (
	InputMode ThreadedIoMode = iota
	OutputMode
)

// ThreadedIoFile implements a thread-safe file I/O wrapper with caching
type ThreadedIoFile struct {
	mu           sync.RWMutex
	internalFile File
	mode         ThreadedIoMode
	cache        *IoCache
	ioBuffer     []byte
	position     uint64
	size         uint64
	eof          int32 // atomic bool
	internalFileError int64 // atomic error code
	
	// Flush synchronization
	flushMutex    sync.Mutex
	flushing      bool
	flushComplete bool
	
	// Task synchronization
	taskExitedMutex sync.Mutex
	taskExited      bool
	
	// Task control
	stopChannel chan bool
	taskWG      sync.WaitGroup
}

// NewThreadedIoFile creates a new threaded I/O file wrapper
func NewThreadedIoFile(internalFile File, mode ThreadedIoMode, ioCacheSize, ioBlockSize uint64) *ThreadedIoFile {
	return &ThreadedIoFile{
		internalFile: internalFile,
		mode:         mode,
		cache:        NewIoCache(ioCacheSize),
		ioBuffer:     make([]byte, ioBlockSize),
		stopChannel:  make(chan bool, 1),
	}
}

// Close closes the threaded I/O file
func (tif *ThreadedIoFile) Close() error {
	tif.mu.Lock()
	defer tif.mu.Unlock()
	
	// Stop the background task
	select {
	case tif.stopChannel <- true:
	default:
	}
	
	// Wait for task to exit
	tif.taskWG.Wait()
	
	// Close the internal file
	if tif.internalFile != nil {
		err := tif.internalFile.Close()
		tif.internalFile = nil
		return err
	}
	
	return nil
}

// Read reads data from the file using the cache
func (tif *ThreadedIoFile) Read(buffer []byte) (int64, error) {
	tif.mu.RLock()
	defer tif.mu.RUnlock()
	
	if tif.mode != InputMode {
		return 0, errors.New("file not opened for reading")
	}
	
	// Check for internal file error
	if errCode := atomic.LoadInt64(&tif.internalFileError); errCode != 0 {
		return 0, errors.New("internal file error")
	}
	
	// Try to read from cache
	bytesRead, err := tif.cache.Read(buffer)
	if err != nil && err != io.EOF {
		return 0, err
	}
	
	// Update position
	tif.position += uint64(bytesRead)
	
	// Check if we've reached EOF
	if atomic.LoadInt32(&tif.eof) != 0 && bytesRead == 0 {
		return 0, io.EOF
	}
	
	return int64(bytesRead), nil
}

// Write writes data to the file using the cache
func (tif *ThreadedIoFile) Write(buffer []byte) (int64, error) {
	tif.mu.RLock()
	defer tif.mu.RUnlock()
	
	if tif.mode != OutputMode {
		return 0, errors.New("file not opened for writing")
	}
	
	// Check for internal file error
	if errCode := atomic.LoadInt64(&tif.internalFileError); errCode != 0 {
		return 0, errors.New("internal file error")
	}
	
	// Write to cache
	bytesWritten, err := tif.cache.Write(buffer)
	if err != nil {
		return 0, err
	}
	
	// Update position
	tif.position += uint64(bytesWritten)
	
	return int64(bytesWritten), nil
}

// CloseForWriting signals that no more writes will occur
func (tif *ThreadedIoFile) CloseForWriting() {
	if tif.mode == OutputMode {
		tif.cache.CloseForWriting()
	}
}

// Size returns the size of the file
func (tif *ThreadedIoFile) Size() (int64, error) {
	tif.mu.RLock()
	defer tif.mu.RUnlock()
	
	if tif.internalFile == nil {
		return 0, errors.New("file not opened")
	}
	
	return tif.internalFile.Size()
}

// Flush flushes any cached data to the internal file
func (tif *ThreadedIoFile) Flush() error {
	tif.flushMutex.Lock()
	defer tif.flushMutex.Unlock()
	
	if tif.mode != OutputMode {
		return nil // Nothing to flush for input mode
	}
	
	tif.flushing = true
	tif.flushComplete = false
	
	// Signal cache to flush
	err := tif.cache.Flush()
	if err != nil {
		tif.flushing = false
		return err
	}
	
	// Wait for flush to complete
	for !tif.flushComplete && !tif.taskExited {
		time.Sleep(1 * time.Millisecond)
	}
	
	tif.flushing = false
	return nil
}

// Seek seeks to a position in the file
func (tif *ThreadedIoFile) Seek(position uint64) error {
	return errors.New("ThreadedIoFile does not support Seek")
}

// Tell returns the current position in the file
func (tif *ThreadedIoFile) Tell() (uint64, error) {
	tif.mu.RLock()
	defer tif.mu.RUnlock()
	
	return tif.position, nil
}

// Open opens the threaded I/O file
func (tif *ThreadedIoFile) Open() error {
	tif.mu.Lock()
	defer tif.mu.Unlock()
	
	if tif.internalFile == nil {
		return errors.New("internal file not set")
	}
	
	// Open the internal file
	err := tif.internalFile.Open()
	if err != nil {
		return err
	}
	
	// Get file size for input mode
	if tif.mode == InputMode {
		size, err := tif.internalFile.Size()
		if err == nil && size >= 0 {
			tif.size = uint64(size)
		}
	}
	
	// Start the background task
	tif.taskWG.Add(1)
	go tif.taskHandler()
	
	return nil
}

// taskHandler is the main background task handler
func (tif *ThreadedIoFile) taskHandler() {
	defer tif.taskWG.Done()
	defer func() {
		tif.taskExitedMutex.Lock()
		tif.taskExited = true
		tif.taskExitedMutex.Unlock()
	}()
	
	switch tif.mode {
	case InputMode:
		tif.runInInputMode()
	case OutputMode:
		tif.runInOutputMode()
	}
}

// runInInputMode handles input mode operations
func (tif *ThreadedIoFile) runInInputMode() {
	for {
		select {
		case <-tif.stopChannel:
			return
		default:
			// Check if cache needs data
			if tif.cache.BytesAvailable() < uint64(len(tif.ioBuffer)) {
				// Read from internal file
				bytesRead, err := tif.internalFile.Read(tif.ioBuffer)
				if err != nil && err != io.EOF {
					atomic.StoreInt64(&tif.internalFileError, 1)
					return
				}
				
				if bytesRead > 0 {
					// Write to cache
					_, cacheErr := tif.cache.Write(tif.ioBuffer[:bytesRead])
					if cacheErr != nil {
						atomic.StoreInt64(&tif.internalFileError, 1)
						return
					}
				}
				
				if err == io.EOF {
					atomic.StoreInt32(&tif.eof, 1)
					tif.cache.CloseForWriting()
					return
				}
			} else {
				// Give other goroutines a chance
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}

// runInOutputMode handles output mode operations  
func (tif *ThreadedIoFile) runInOutputMode() {
	for {
		select {
		case <-tif.stopChannel:
			return
		default:
			// Check if cache has data to write
			bytesRead, err := tif.cache.Read(tif.ioBuffer)
			if err == io.EOF {
				// Cache is closed and empty
				tif.flushMutex.Lock()
				tif.flushComplete = true
				tif.flushMutex.Unlock()
				return
			} else if err != nil {
				atomic.StoreInt64(&tif.internalFileError, 1)
				return
			}
			
			if bytesRead > 0 {
				// Write to internal file
				_, writeErr := tif.internalFile.Write(tif.ioBuffer[:bytesRead])
				if writeErr != nil {
					atomic.StoreInt64(&tif.internalFileError, 1)
					return
				}
			} else {
				// Give other goroutines a chance
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}