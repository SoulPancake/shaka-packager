// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"log"
)

// FileCloser provides automatic file closing functionality 
// Used with defer statements to ensure files are properly closed
type FileCloser struct {
	File File
}

// NewFileCloser creates a new file closer wrapper
func NewFileCloser(file File) *FileCloser {
	return &FileCloser{File: file}
}

// Close closes the wrapped file and logs any errors
func (fc *FileCloser) Close() error {
	if fc.File != nil {
		err := fc.File.Close()
		if err != nil {
			log.Printf("Failed to close file properly: %v", err)
			return err
		}
	}
	return nil
}

// AutoClose is a helper function to be used with defer for automatic file closing
// Example usage: defer file.AutoClose(myFile)
func AutoClose(file File) {
	if file != nil {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close file properly: %v", err)
		}
	}
}

// WithAutoClose wraps a file operation with automatic closing
// The provided function will be called with the file, and the file will be
// automatically closed when the function returns
func WithAutoClose(file File, fn func(File) error) error {
	if file == nil {
		return nil
	}
	
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file properly: %v", err)
		}
	}()
	
	return fn(file)
}