// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package file provides file I/O implementations for the shaka packager.
package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

// File prefixes
const (
	CallbackFilePrefix = "callback://"
	LocalFilePrefix    = "file://"
	MemoryFilePrefix   = "memory://"
	UdpFilePrefix      = "udp://"
	HttpFilePrefix     = "http://"
	HttpsFilePrefix    = "https://"
	WholeFile          = -1
)

// Configuration flags
var (
	IOCacheSize  uint64 = 32 << 20 // 32MB
	IOBlockSize  uint64 = 1 << 16  // 64KB
)

// FileInterface defines the interface for all file types.
type FileInterface interface {
	Close() bool
	Read(buffer []byte) (int64, error)
	Write(buffer []byte) (int64, error)
	CloseForWriting()
	Size() int64
	Flush() bool
	Seek(position uint64) bool
	Tell() (uint64, bool)
	FileName() string
}

// BaseFile provides common functionality for all file implementations.
type BaseFile struct {
	fileName string
	mu       sync.RWMutex
}

// NewBaseFile creates a new BaseFile.
func NewBaseFile(fileName string) *BaseFile {
	return &BaseFile{fileName: fileName}
}

// FileName returns the file name with prefix stripped.
func (f *BaseFile) FileName() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fileName
}

// FileFactory provides methods to create different types of files.
type FileFactory struct{}

// Open opens a file based on its prefix.
func (ff *FileFactory) Open(fileName, mode string) (FileInterface, packager.Status) {
	if strings.HasPrefix(fileName, LocalFilePrefix) || 
	   (!strings.Contains(fileName, "://")) {
		return ff.openLocalFile(fileName, mode)
	} else if strings.HasPrefix(fileName, MemoryFilePrefix) {
		return ff.openMemoryFile(fileName, mode)
	} else if strings.HasPrefix(fileName, UdpFilePrefix) {
		return ff.openUdpFile(fileName, mode)
	} else if strings.HasPrefix(fileName, HttpFilePrefix) ||
	          strings.HasPrefix(fileName, HttpsFilePrefix) {
		return ff.openHttpFile(fileName, mode)
	} else if strings.HasPrefix(fileName, CallbackFilePrefix) {
		return ff.openCallbackFile(fileName, mode)
	}
	
	return nil, packager.NewStatus(packager.INVALID_ARGUMENT, 
		fmt.Sprintf("Unsupported file type: %s", fileName))
}

// OpenWithNoBuffering opens a file with no buffering.
func (ff *FileFactory) OpenWithNoBuffering(fileName, mode string) (FileInterface, packager.Status) {
	// For now, same as regular open - buffering can be controlled per implementation
	return ff.Open(fileName, mode)
}

func (ff *FileFactory) openLocalFile(fileName, mode string) (FileInterface, packager.Status) {
	return NewLocalFile(fileName, mode)
}

func (ff *FileFactory) openMemoryFile(fileName, mode string) (FileInterface, packager.Status) {
	return NewMemoryFile(fileName, mode)
}

func (ff *FileFactory) openUdpFile(fileName, mode string) (FileInterface, packager.Status) {
	return NewUdpFile(fileName, mode)
}

func (ff *FileFactory) openHttpFile(fileName, mode string) (FileInterface, packager.Status) {
	return NewHttpFile(fileName, mode)
}

func (ff *FileFactory) openCallbackFile(fileName, mode string) (FileInterface, packager.Status) {
	return NewCallbackFile(fileName, mode)
}

// Delete deletes a file.
func (ff *FileFactory) Delete(fileName string) bool {
	// Strip prefix and delete local file
	actualFileName := fileName
	if strings.HasPrefix(fileName, LocalFilePrefix) {
		actualFileName = fileName[len(LocalFilePrefix):]
	}
	
	err := os.Remove(actualFileName)
	return err == nil
}

// GetFileSize returns the size of a file.
func (ff *FileFactory) GetFileSize(fileName string) int64 {
	actualFileName := fileName
	if strings.HasPrefix(fileName, LocalFilePrefix) {
		actualFileName = fileName[len(LocalFilePrefix):]
	}
	
	info, err := os.Stat(actualFileName)
	if err != nil {
		return -1
	}
	
	return info.Size()
}

// ReadFileToString reads entire file content into a string.
func (ff *FileFactory) ReadFileToString(fileName string) (string, packager.Status) {
	file, status := ff.Open(fileName, "r")
	if !status.OK() {
		return "", status
	}
	defer file.Close()
	
	size := file.Size()
	if size < 0 {
		return "", packager.NewStatus(packager.FILE_FAILURE, "Cannot determine file size")
	}
	
	buffer := make([]byte, size)
	bytesRead, err := file.Read(buffer)
	if err != nil {
		return "", packager.NewStatus(packager.FILE_FAILURE, err.Error())
	}
	
	return string(buffer[:bytesRead]), packager.StatusOK
}

// WriteStringToFile writes a string to a file.
func (ff *FileFactory) WriteStringToFile(fileName, contents string) packager.Status {
	file, status := ff.Open(fileName, "w")
	if !status.OK() {
		return status
	}
	defer file.Close()
	
	_, err := file.Write([]byte(contents))
	if err != nil {
		return packager.NewStatus(packager.FILE_FAILURE, err.Error())
	}
	
	return packager.StatusOK
}

// WriteFileAtomically writes content to a file atomically.
func (ff *FileFactory) WriteFileAtomically(fileName, contents string) packager.Status {
	tempFile := fileName + ".tmp"
	
	status := ff.WriteStringToFile(tempFile, contents)
	if !status.OK() {
		ff.Delete(tempFile)
		return status
	}
	
	actualFileName := fileName
	actualTempFile := tempFile
	
	if strings.HasPrefix(fileName, LocalFilePrefix) {
		actualFileName = fileName[len(LocalFilePrefix):]
		actualTempFile = tempFile[len(LocalFilePrefix):]
	}
	
	err := os.Rename(actualTempFile, actualFileName)
	if err != nil {
		ff.Delete(tempFile)
		return packager.NewStatus(packager.FILE_FAILURE, err.Error())
	}
	
	return packager.StatusOK
}

// Copy copies files.
func (ff *FileFactory) Copy(fromFileName, toFileName string) packager.Status {
	fromFile, status := ff.Open(fromFileName, "r")
	if !status.OK() {
		return status
	}
	defer fromFile.Close()
	
	toFile, status := ff.Open(toFileName, "w")
	if !status.OK() {
		return status
	}
	defer toFile.Close()
	
	_, status2 := ff.CopyFiles(fromFile, toFile, WholeFile)
	return status2
}

// CopyFiles copies from one file to another.
func (ff *FileFactory) CopyFiles(source, destination FileInterface, maxCopy int64) (int64, packager.Status) {
	const bufferSize = 64 * 1024
	buffer := make([]byte, bufferSize)
	totalCopied := int64(0)
	
	for {
		if maxCopy >= 0 && totalCopied >= maxCopy {
			break
		}
		
		readSize := bufferSize
		if maxCopy >= 0 && totalCopied+int64(readSize) > maxCopy {
			readSize = int(maxCopy - totalCopied)
		}
		
		bytesRead, err := source.Read(buffer[:readSize])
		if err != nil {
			if err == io.EOF {
				break
			}
			return totalCopied, packager.NewStatus(packager.FILE_FAILURE, err.Error())
		}
		
		if bytesRead == 0 {
			break
		}
		
		bytesWritten, err := destination.Write(buffer[:bytesRead])
		if err != nil {
			return totalCopied, packager.NewStatus(packager.FILE_FAILURE, err.Error())
		}
		
		totalCopied += bytesWritten
	}
	
	return totalCopied, packager.StatusOK
}

// IsLocalRegularFile checks if a file is a local regular file.
func (ff *FileFactory) IsLocalRegularFile(fileName string) bool {
	actualFileName := fileName
	if strings.HasPrefix(fileName, LocalFilePrefix) {
		actualFileName = fileName[len(LocalFilePrefix):]
	}
	
	// Check if it's not a URL scheme
	if strings.Contains(actualFileName, "://") {
		return false
	}
	
	info, err := os.Stat(actualFileName)
	if err != nil {
		return false
	}
	
	return info.Mode().IsRegular()
}

// MakeCallbackFileName creates a callback file name.
func (ff *FileFactory) MakeCallbackFileName(callbackParams packager.BufferCallbackParams, name string) string {
	// Create a unique identifier for the callback
	// In the real implementation, this would embed the callback params
	return fmt.Sprintf("%s%p_%s", CallbackFilePrefix, &callbackParams, name)
}

// ParseCallbackFileName parses a callback file name.
func (ff *FileFactory) ParseCallbackFileName(callbackFileName string) (*packager.BufferCallbackParams, string, packager.Status) {
	if !strings.HasPrefix(callbackFileName, CallbackFilePrefix) {
		return nil, "", packager.NewStatus(packager.INVALID_ARGUMENT, "Not a callback file name")
	}
	
	remainder := callbackFileName[len(CallbackFilePrefix):]
	parts := strings.SplitN(remainder, "_", 2)
	if len(parts) != 2 {
		return nil, "", packager.NewStatus(packager.INVALID_ARGUMENT, "Invalid callback file name format")
	}
	
	// In a real implementation, we'd reconstruct the callback params from the embedded data
	// For now, return nil callback params and the name
	return nil, parts[1], packager.StatusOK
}

// Global file factory instance
var GlobalFileFactory = &FileFactory{}