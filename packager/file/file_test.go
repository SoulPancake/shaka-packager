// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"os"
	"strings"
	"testing"
	"github.com/SoulPancake/shaka-packager/include/packager"
)

func TestLocalFileReadWrite(t *testing.T) {
	tempFile := "/tmp/test_local_file.txt"
	defer os.Remove(tempFile)

	// Test writing
	file, status := NewLocalFile(tempFile, "w")
	if !status.OK() {
		t.Fatalf("Failed to create local file: %s", status.String())
	}

	testData := "Hello, World!"
	bytesWritten, err := file.Write([]byte(testData))
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	if bytesWritten != int64(len(testData)) {
		t.Fatalf("Expected %d bytes written, got %d", len(testData), bytesWritten)
	}

	if !file.Close() {
		t.Fatal("Failed to close file")
	}

	// Test reading
	file, status = NewLocalFile(tempFile, "r")
	if !status.OK() {
		t.Fatalf("Failed to open local file for reading: %s", status.String())
	}

	buffer := make([]byte, 100)
	bytesRead, err := file.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from file: %v", err)
	}

	if bytesRead != int64(len(testData)) {
		t.Fatalf("Expected %d bytes read, got %d", len(testData), bytesRead)
	}

	if string(buffer[:bytesRead]) != testData {
		t.Fatalf("Expected '%s', got '%s'", testData, string(buffer[:bytesRead]))
	}

	if !file.Close() {
		t.Fatal("Failed to close file")
	}
}

func TestMemoryFile(t *testing.T) {
	file, status := NewMemoryFile("memory://test", "w+")
	if !status.OK() {
		t.Fatalf("Failed to create memory file: %s", status.String())
	}

	testData := "Memory file test data"
	bytesWritten, err := file.Write([]byte(testData))
	if err != nil {
		t.Fatalf("Failed to write to memory file: %v", err)
	}

	if bytesWritten != int64(len(testData)) {
		t.Fatalf("Expected %d bytes written, got %d", len(testData), bytesWritten)
	}

	// Test size
	size := file.Size()
	if size != int64(len(testData)) {
		t.Fatalf("Expected size %d, got %d", len(testData), size)
	}

	// Test seek
	if !file.Seek(0) {
		t.Fatal("Failed to seek to beginning")
	}

	// Test read
	buffer := make([]byte, 100)
	bytesRead, err := file.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from memory file: %v", err)
	}

	if bytesRead != int64(len(testData)) {
		t.Fatalf("Expected %d bytes read, got %d", len(testData), bytesRead)
	}

	if string(buffer[:bytesRead]) != testData {
		t.Fatalf("Expected '%s', got '%s'", testData, string(buffer[:bytesRead]))
	}

	if !file.Close() {
		t.Fatal("Failed to close memory file")
	}
}

func TestFileFactory(t *testing.T) {
	factory := &FileFactory{}

	// Test local file creation
	tempFile := "/tmp/test_factory.txt"
	defer os.Remove(tempFile)

	file, status := factory.Open(tempFile, "w")
	if !status.OK() {
		t.Fatalf("Failed to open local file via factory: %s", status.String())
	}

	if !strings.Contains(file.FileName(), "test_factory.txt") {
		t.Fatalf("Unexpected file name: %s", file.FileName())
	}

	file.Close()

	// Test memory file creation
	memFile, status := factory.Open("memory://test_mem", "w")
	if !status.OK() {
		t.Fatalf("Failed to open memory file via factory: %s", status.String())
	}

	if memFile.FileName() != "test_mem" {
		t.Fatalf("Expected 'test_mem', got '%s'", memFile.FileName())
	}

	memFile.Close()
}

func TestFileFactoryOperations(t *testing.T) {
	factory := &FileFactory{}
	tempFile := "/tmp/test_factory_ops.txt"
	defer os.Remove(tempFile)

	// Test WriteStringToFile
	testContent := "Factory test content"
	status := factory.WriteStringToFile(tempFile, testContent)
	if !status.OK() {
		t.Fatalf("Failed to write string to file: %s", status.String())
	}

	// Test ReadFileToString
	content, status := factory.ReadFileToString(tempFile)
	if !status.OK() {
		t.Fatalf("Failed to read file to string: %s", status.String())
	}

	if content != testContent {
		t.Fatalf("Expected '%s', got '%s'", testContent, content)
	}

	// Test GetFileSize
	size := factory.GetFileSize(tempFile)
	if size != int64(len(testContent)) {
		t.Fatalf("Expected size %d, got %d", len(testContent), size)
	}

	// Test IsLocalRegularFile
	if !factory.IsLocalRegularFile(tempFile) {
		t.Fatal("Expected file to be identified as local regular file")
	}

	// Test Delete
	if !factory.Delete(tempFile) {
		t.Fatal("Failed to delete file")
	}

	// Verify file is deleted
	if factory.IsLocalRegularFile(tempFile) {
		t.Fatal("File should have been deleted")
	}
}

func TestFileCopy(t *testing.T) {
	factory := &FileFactory{}
	sourceFile := "/tmp/test_copy_source.txt"
	destFile := "/tmp/test_copy_dest.txt"
	defer os.Remove(sourceFile)
	defer os.Remove(destFile)

	// Create source file
	testContent := "Copy test content"
	status := factory.WriteStringToFile(sourceFile, testContent)
	if !status.OK() {
		t.Fatalf("Failed to create source file: %s", status.String())
	}

	// Test copy
	status = factory.Copy(sourceFile, destFile)
	if !status.OK() {
		t.Fatalf("Failed to copy file: %s", status.String())
	}

	// Verify destination content
	content, status := factory.ReadFileToString(destFile)
	if !status.OK() {
		t.Fatalf("Failed to read destination file: %s", status.String())
	}

	if content != testContent {
		t.Fatalf("Expected '%s', got '%s'", testContent, content)
	}
}

func TestStatusIntegration(t *testing.T) {
	// Test that status codes work correctly
	status := packager.NewStatus(packager.FILE_FAILURE, "Test error")
	if status.OK() {
		t.Fatal("Status should not be OK")
	}

	if status.ErrorCode() != packager.FILE_FAILURE {
		t.Fatalf("Expected FILE_FAILURE, got %d", status.ErrorCode())
	}

	if status.ErrorMessage() != "Test error" {
		t.Fatalf("Expected 'Test error', got '%s'", status.ErrorMessage())
	}

	// Test status update
	newStatus := packager.NewStatus(packager.INVALID_ARGUMENT, "Another error")
	status.Update(newStatus) // Should not update since status is not OK

	if status.ErrorCode() != packager.FILE_FAILURE {
		t.Fatal("Status should not have been updated")
	}

	okStatus := packager.StatusOK
	okStatus.Update(newStatus) // Should update since okStatus is OK

	if okStatus.ErrorCode() != packager.INVALID_ARGUMENT {
		t.Fatal("Status should have been updated")
	}
}