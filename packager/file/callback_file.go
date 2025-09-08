// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

// CallbackFile implements File interface, which delegates read/write calls 
// to the callback functions set through the file name.
type CallbackFile struct {
	mu             sync.RWMutex
	fileName       string
	fileMode       string
	name           string
	callbackParams *packager.BufferCallbackParams
	closed         bool
}

// NewCallbackFile creates a new callback file
// fileName is the callback file name, which should have callback address encoded.
// Note that the file type prefix should be stripped off already.
// mode is a string containing a file access mode, refer to standard file modes.
func NewCallbackFile(fileName, mode string) *CallbackFile {
	return &CallbackFile{
		fileName: fileName,
		fileMode: mode,
	}
}

// Close closes the file
func (cf *CallbackFile) Close() error {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	
	if cf.closed {
		return nil
	}
	cf.closed = true
	return nil
}

// Read reads data from the file
func (cf *CallbackFile) Read(buffer []byte) (int64, error) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	
	if cf.closed {
		return 0, errors.New("file is closed")
	}
	
	if cf.callbackParams == nil || cf.callbackParams.ReadFunc == nil {
		log.Printf("Read function not defined")
		return 0, errors.New("read function not defined")
	}
	
	bytesRead := cf.callbackParams.ReadFunc(cf.name, buffer, uint64(len(buffer)))
	if bytesRead < 0 {
		return 0, errors.New("read error")
	}
	return bytesRead, nil
}

// Write writes data to the file
func (cf *CallbackFile) Write(buffer []byte) (int64, error) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	
	if cf.closed {
		return 0, errors.New("file is closed")
	}
	
	if cf.callbackParams == nil || cf.callbackParams.WriteFunc == nil {
		log.Printf("Write function not defined")
		return 0, errors.New("write function not defined")
	}
	
	bytesWritten := cf.callbackParams.WriteFunc(cf.name, buffer, uint64(len(buffer)))
	if bytesWritten < 0 {
		return 0, errors.New("write error")
	}
	return bytesWritten, nil
}

// CloseForWriting closes the file for writing
func (cf *CallbackFile) CloseForWriting() {
	// Do nothing for callback file
}

// Size returns the size of the file
func (cf *CallbackFile) Size() (int64, error) {
	log.Printf("CallbackFile does not support Size()")
	return -1, errors.New("size not supported")
}

// Flush flushes the file
func (cf *CallbackFile) Flush() error {
	// Do nothing on Flush for callback file
	return nil
}

// Seek seeks to a position in the file
func (cf *CallbackFile) Seek(position uint64) error {
	log.Printf("CallbackFile does not support Seek()")
	return errors.New("seek not supported")
}

// Tell returns the current position in the file
func (cf *CallbackFile) Tell() (uint64, error) {
	log.Printf("CallbackFile does not support Tell()")
	return 0, errors.New("tell not supported")
}

// Open opens the file
func (cf *CallbackFile) Open() error {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	
	if cf.fileMode != "r" && cf.fileMode != "w" && cf.fileMode != "rb" && cf.fileMode != "wb" {
		return fmt.Errorf("CallbackFile does not support file mode %s", cf.fileMode)
	}
	
	callbackParams, name, err := ParseCallbackFileName(cf.fileName)
	if err != nil {
		return err
	}
	
	cf.callbackParams = callbackParams
	cf.name = name
	return nil
}

// ParseCallbackFileName parses callback file name to extract parameters and name
func ParseCallbackFileName(fileName string) (*packager.BufferCallbackParams, string, error) {
	// Implementation would parse the encoded callback parameters from filename
	// For now, return a basic implementation
	return &packager.BufferCallbackParams{}, fileName, nil
}