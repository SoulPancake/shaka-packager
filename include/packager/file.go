// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// File interface constants
const (
	CallbackFilePrefix = "callback://"
	LocalFilePrefix    = "file://"
	MemoryFilePrefix   = "memory://"
	UdpFilePrefix      = "udp://"
	HttpFilePrefix     = "http://"
	WholeFile          = -1
)

// File defines an abstract file interface.
type File interface {
	// Close flushes and de-allocates resources associated with this file.
	// Returns true on success. For writable files, returning false may indicate data loss.
	Close() bool

	// Read reads data into buffer.
	// Returns number of bytes read, or a value < 0 on error.
	// Zero on end-of-file, or if 'length' is zero.
	Read(buffer []byte) (int64, error)

	// Write writes block of data.
	// Returns number of bytes written, or a value < 0 on error.
	Write(buffer []byte) (int64, error)

	// CloseForWriting closes the file for writing. This signals that no more data
	// will be written. Future writes are invalid and their behavior is undefined!
	// Data may still be read from the file after calling this method.
	// Some implementations may ignore this if they cannot use the signal.
	CloseForWriting()

	// Size returns the size of the file in bytes. A return value less than zero
	// indicates a problem getting the size.
	Size() int64

	// Flush flushes the file so that recently written data will survive an
	// application crash (but not necessarily an OS crash). For instance, in
	// LocalFile the data is flushed into the OS but not necessarily to disk.
	Flush() bool

	// Seek seeks to the specified position in the file.
	Seek(position uint64) bool

	// Tell gets the current file position.
	Tell() (uint64, bool)

	// FileName returns the file name. Note that the file type prefix has been stripped off.
	FileName() string
}

// FileUtils provides utility functions for file operations.
type FileUtils struct{}

// Open opens the specified file.
// This is a file factory method, it opens a proper file automatically
// based on prefix, e.g. "file://" for LocalFile.
func (fu FileUtils) Open(fileName, mode string) (File, error) {
	// TODO: Implement file opening logic based on prefix
	return nil, NewStatus(UNIMPLEMENTED, "File opening not implemented yet")
}

// OpenWithNoBuffering opens the specified file in direct-access mode (no buffering).
func (fu FileUtils) OpenWithNoBuffering(fileName, mode string) (File, error) {
	// TODO: Implement unbuffered file opening
	return nil, NewStatus(UNIMPLEMENTED, "Unbuffered file opening not implemented yet")
}

// Delete deletes the specified file.
func (fu FileUtils) Delete(fileName string) bool {
	// TODO: Implement file deletion
	return false
}

// GetFileSize returns the size of a file in bytes on success, a value < 0 otherwise.
// The file will be opened and closed in the process.
func (fu FileUtils) GetFileSize(fileName string) int64 {
	// TODO: Implement file size retrieval
	return -1
}

// ReadFileToString reads the contents of a file into string.
func (fu FileUtils) ReadFileToString(fileName string) (string, error) {
	// TODO: Implement file reading to string
	return "", NewStatus(UNIMPLEMENTED, "ReadFileToString not implemented yet")
}

// WriteStringToFile writes the data to file.
func (fu FileUtils) WriteStringToFile(fileName, contents string) error {
	// TODO: Implement string writing to file
	return NewStatus(UNIMPLEMENTED, "WriteStringToFile not implemented yet")
}

// WriteFileAtomically saves contents to fileName in an atomic manner.
func (fu FileUtils) WriteFileAtomically(fileName, contents string) error {
	// TODO: Implement atomic file writing
	return NewStatus(UNIMPLEMENTED, "WriteFileAtomically not implemented yet")
}

// Copy copies files. This is not good for copying huge files. Although not
// recommended, it is safe to have source file and destination file name be the same.
func (fu FileUtils) Copy(fromFileName, toFileName string) error {
	// TODO: Implement file copying
	return NewStatus(UNIMPLEMENTED, "File copy not implemented yet")
}

// CopyFiles copies the contents from source to destination.
func (fu FileUtils) CopyFiles(source, destination File) (int64, error) {
	// TODO: Implement file-to-file copying
	return 0, NewStatus(UNIMPLEMENTED, "File copying not implemented yet")
}

// CopyFilesWithLimit copies the contents from source to destination with a maximum limit.
func (fu FileUtils) CopyFilesWithLimit(source, destination File, maxCopy int64) (int64, error) {
	// TODO: Implement limited file copying
	return 0, NewStatus(UNIMPLEMENTED, "Limited file copying not implemented yet")
}

// IsLocalRegularFile returns true if fileName is a local and regular file.
func (fu FileUtils) IsLocalRegularFile(fileName string) bool {
	// TODO: Implement local file checking
	return false
}

// MakeCallbackFileName generates callback file name.
// NOTE: THE GENERATED NAME IS ONLY VALID WHILE callbackParams IS VALID.
func (fu FileUtils) MakeCallbackFileName(callbackParams BufferCallbackParams, name string) string {
	// TODO: Implement callback file name generation
	return ""
}

// ParseCallbackFileName parses and extracts callback params.
func (fu FileUtils) ParseCallbackFileName(callbackFileName string) (*BufferCallbackParams, string, error) {
	// TODO: Implement callback file name parsing
	return nil, "", NewStatus(UNIMPLEMENTED, "ParseCallbackFileName not implemented yet")
}

// Global file utils instance
var GlobalFileUtils = FileUtils{}