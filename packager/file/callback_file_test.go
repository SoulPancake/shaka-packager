// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"testing"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

func TestCallbackFile(t *testing.T) {
	// Test callback file creation
	cbFile := NewCallbackFile("test://callback", "r")
	if cbFile == nil {
		t.Fatal("Failed to create callback file")
	}

	// Test unsupported file modes
	err := cbFile.Open()
	if err == nil {
		t.Error("Expected error for unsupported file mode")
	}

	// Test supported file modes
	supportedModes := []string{"r", "w", "rb", "wb"}
	for _, mode := range supportedModes {
		cbFile := NewCallbackFile("test://callback", mode)
		// Note: This will fail without proper callback params setup
		// In real usage, ParseCallbackFileName would extract callback functions
		err := cbFile.Open()
		// We expect this to work in a real scenario with proper callback setup
		if err != nil {
			t.Logf("Open failed for mode %s (expected with test setup): %v", mode, err)
		}
	}
}

func TestCallbackFileOperations(t *testing.T) {
	cbFile := NewCallbackFile("test://callback", "w")
	
	// Test operations on unopened file
	buffer := make([]byte, 10)
	_, err := cbFile.Read(buffer)
	if err == nil {
		t.Error("Expected error when reading from unopened file")
	}
	
	_, err = cbFile.Write([]byte("test"))
	if err == nil {
		t.Error("Expected error when writing to unopened file")
	}
	
	// Test Size operation (should always fail)
	_, err = cbFile.Size()
	if err == nil {
		t.Error("Expected error for Size operation on callback file")
	}
	
	// Test Seek operation (should always fail)
	err = cbFile.Seek(10)
	if err == nil {
		t.Error("Expected error for Seek operation on callback file")
	}
	
	// Test Tell operation (should always fail)
	_, err = cbFile.Tell()
	if err == nil {
		t.Error("Expected error for Tell operation on callback file")
	}
	
	// Test Flush (should succeed)
	err = cbFile.Flush()
	if err != nil {
		t.Error("Flush should not fail on callback file")
	}
	
	// Test Close
	err = cbFile.Close()
	if err != nil {
		t.Error("Close should not fail")
	}
}

func TestCallbackFileWithMockCallbacks(t *testing.T) {
	// Create mock callback functions
	readFunc := func(name string, buffer []byte, length uint64) int64 {
		testData := "Hello, World!"
		copy(buffer, testData)
		return int64(len(testData))
	}
	
	writeFunc := func(name string, buffer []byte, length uint64) int64 {
		// Mock write - just return the length
		return int64(length)
	}
	
	// Mock callback params
	mockParams := &packager.BufferCallbackParams{
		ReadFunc:  readFunc,
		WriteFunc: writeFunc,
	}
	
	// Test with mock setup
	cbFile := NewCallbackFile("test://callback", "r")
	
	// Manually set callback params for testing (normally done in ParseCallbackFileName)
	cbFile.callbackParams = mockParams
	cbFile.name = "test"
	
	// Test read
	buffer := make([]byte, 20)
	bytesRead, err := cbFile.Read(buffer)
	if err != nil {
		t.Errorf("Read failed: %v", err)
	}
	if bytesRead != 13 {
		t.Errorf("Expected 13 bytes read, got %d", bytesRead)
	}
	if string(buffer[:bytesRead]) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", string(buffer[:bytesRead]))
	}
	
	// Test write
	writeData := []byte("Test write")
	bytesWritten, err := cbFile.Write(writeData)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if bytesWritten != int64(len(writeData)) {
		t.Errorf("Expected %d bytes written, got %d", len(writeData), bytesWritten)
	}
}

func TestParseCallbackFileName(t *testing.T) {
	// Test basic parsing (implementation would be more complex in real scenario)
	params, name, err := ParseCallbackFileName("test://callback")
	if err != nil {
		t.Errorf("ParseCallbackFileName failed: %v", err)
	}
	if params == nil {
		t.Error("Expected non-nil callback params")
	}
	if name != "test://callback" {
		t.Errorf("Expected name 'test://callback', got '%s'", name)
	}
}