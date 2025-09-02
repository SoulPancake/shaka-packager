// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

// LocalFile represents a local file system file.
type LocalFile struct {
	*BaseFile
	file     *os.File
	mode     string
	mu       sync.RWMutex
	position uint64
}

// NewLocalFile creates a new local file.
func NewLocalFile(fileName, mode string) (*LocalFile, packager.Status) {
	// Strip file:// prefix if present
	actualFileName := fileName
	if strings.HasPrefix(fileName, LocalFilePrefix) {
		actualFileName = fileName[len(LocalFilePrefix):]
	}

	f := &LocalFile{
		BaseFile: NewBaseFile(actualFileName),
		mode:     mode,
	}

	status := f.open()
	if !status.OK() {
		return nil, status
	}

	return f, packager.StatusOK
}

// open opens the local file.
func (lf *LocalFile) open() packager.Status {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	var flag int
	switch lf.mode {
	case "r":
		flag = os.O_RDONLY
	case "w":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "a":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "r+":
		flag = os.O_RDWR
	case "w+":
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a+":
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		return packager.NewStatus(packager.INVALID_ARGUMENT, "Invalid file mode: "+lf.mode)
	}

	var err error
	lf.file, err = os.OpenFile(lf.fileName, flag, 0644)
	if err != nil {
		return packager.NewStatus(packager.FILE_FAILURE, err.Error())
	}

	return packager.StatusOK
}

// Close closes the file.
func (lf *LocalFile) Close() bool {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.file == nil {
		return true
	}

	err := lf.file.Close()
	lf.file = nil
	return err == nil
}

// Read reads data from the file.
func (lf *LocalFile) Read(buffer []byte) (int64, error) {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.file == nil {
		return 0, packager.NewStatus(packager.FILE_FAILURE, "File not open").Error()
	}

	n, err := lf.file.Read(buffer)
	if n > 0 {
		lf.position += uint64(n)
	}

	if err == io.EOF {
		return int64(n), io.EOF
	}

	return int64(n), err
}

// Write writes data to the file.
func (lf *LocalFile) Write(buffer []byte) (int64, error) {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.file == nil {
		return 0, packager.NewStatus(packager.FILE_FAILURE, "File not open").Error()
	}

	n, err := lf.file.Write(buffer)
	if n > 0 {
		lf.position += uint64(n)
	}

	return int64(n), err
}

// CloseForWriting closes the file for writing.
func (lf *LocalFile) CloseForWriting() {
	// For local files, this is the same as flush
	lf.Flush()
}

// Size returns the file size.
func (lf *LocalFile) Size() int64 {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	if lf.file == nil {
		return -1
	}

	info, err := lf.file.Stat()
	if err != nil {
		return -1
	}

	return info.Size()
}

// Flush flushes the file.
func (lf *LocalFile) Flush() bool {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.file == nil {
		return false
	}

	// Sync ensures data is written to storage
	err := lf.file.Sync()
	return err == nil
}

// Seek seeks to a position in the file.
func (lf *LocalFile) Seek(position uint64) bool {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.file == nil {
		return false
	}

	offset, err := lf.file.Seek(int64(position), io.SeekStart)
	if err != nil {
		return false
	}

	lf.position = uint64(offset)
	return true
}

// Tell returns the current file position.
func (lf *LocalFile) Tell() (uint64, bool) {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	if lf.file == nil {
		return 0, false
	}

	offset, err := lf.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, false
	}

	lf.position = uint64(offset)
	return lf.position, true
}