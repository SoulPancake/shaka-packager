// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package status

import (
	"strings"
	"testing"
)

func TestStatusBasics(t *testing.T) {
	// Test OK status
	okStatus := StatusOK
	if !okStatus.OK() {
		t.Fatal("StatusOK should be OK")
	}

	if okStatus.ErrorCode() != OK {
		t.Fatalf("Expected OK, got %d", okStatus.ErrorCode())
	}

	if okStatus.ErrorMessage() != "" {
		t.Fatalf("Expected empty message, got '%s'", okStatus.ErrorMessage())
	}

	if okStatus.String() != "OK" {
		t.Fatalf("Expected 'OK', got '%s'", okStatus.String())
	}

	// Test error status
	errorStatus := New(FILE_FAILURE, "File not found")
	if errorStatus.OK() {
		t.Fatal("Error status should not be OK")
	}

	if errorStatus.ErrorCode() != FILE_FAILURE {
		t.Fatalf("Expected FILE_FAILURE, got %d", errorStatus.ErrorCode())
	}

	if errorStatus.ErrorMessage() != "File not found" {
		t.Fatalf("Expected 'File not found', got '%s'", errorStatus.ErrorMessage())
	}

	expectedStr := "3 (FILE_FAILURE): File not found"
	if errorStatus.String() != expectedStr {
		t.Fatalf("Expected '%s', got '%s'", expectedStr, errorStatus.String())
	}
}

func TestStatusUpdate(t *testing.T) {
	status := StatusOK
	
	// Update with error - should update since status is OK
	errorStatus := New(INVALID_ARGUMENT, "Invalid input")
	status.Update(errorStatus)
	
	if status.OK() {
		t.Fatal("Status should not be OK after update")
	}
	
	if status.ErrorCode() != INVALID_ARGUMENT {
		t.Fatalf("Expected INVALID_ARGUMENT, got %d", status.ErrorCode())
	}
	
	// Update again - should not change since status is no longer OK
	anotherError := New(ENCRYPTION_FAILURE, "Encryption failed")
	status.Update(anotherError)
	
	if status.ErrorCode() != INVALID_ARGUMENT {
		t.Fatal("Status should not have been updated the second time")
	}
}

func TestStatusEqual(t *testing.T) {
	status1 := New(FILE_FAILURE, "Error message")
	status2 := New(FILE_FAILURE, "Error message")
	status3 := New(FILE_FAILURE, "Different message")
	status4 := New(INVALID_ARGUMENT, "Error message")
	
	if !status1.Equal(status2) {
		t.Fatal("status1 should equal status2")
	}
	
	if status1.Equal(status3) {
		t.Fatal("status1 should not equal status3 (different message)")
	}
	
	if status1.Equal(status4) {
		t.Fatal("status1 should not equal status4 (different code)")
	}
	
	if !StatusOK.Equal(StatusOK) {
		t.Fatal("StatusOK should equal itself")
	}
}

func TestCodeString(t *testing.T) {
	testCases := []struct {
		code     Code
		expected string
	}{
		{OK, "OK"},
		{UNKNOWN, "UNKNOWN"},
		{CANCELLED, "CANCELLED"},
		{INVALID_ARGUMENT, "INVALID_ARGUMENT"},
		{FILE_FAILURE, "FILE_FAILURE"},
		{END_OF_STREAM, "END_OF_STREAM"},
		{HTTP_FAILURE, "HTTP_FAILURE"},
		{PARSER_FAILURE, "PARSER_FAILURE"},
		{ENCRYPTION_FAILURE, "ENCRYPTION_FAILURE"},
		{MUXER_FAILURE, "MUXER_FAILURE"},
		{INTERNAL_ERROR, "INTERNAL_ERROR"},
	}
	
	for _, tc := range testCases {
		if tc.code.String() != tc.expected {
			t.Fatalf("Code %d: expected '%s', got '%s'", int(tc.code), tc.expected, tc.code.String())
		}
	}
	
	// Test unknown code
	unknownCode := Code(999)
	if unknownCode.String() != "UNKNOWN_STATUS" {
		t.Fatalf("Unknown code should return 'UNKNOWN_STATUS', got '%s'", unknownCode.String())
	}
}

func TestStatusAsError(t *testing.T) {
	// Test OK status as error
	okStatus := StatusOK
	if okStatus.Error() != "" {
		t.Fatal("OK status should have empty error string")
	}
	
	// Test error status as error
	errorStatus := New(FILE_FAILURE, "Test error")
	errorStr := errorStatus.Error()
	
	if errorStr == "" {
		t.Fatal("Error status should have non-empty error string")
	}
	
	if !strings.Contains(errorStr, "FILE_FAILURE") {
		t.Fatalf("Error string should contain 'FILE_FAILURE', got '%s'", errorStr)
	}
	
	if !strings.Contains(errorStr, "Test error") {
		t.Fatalf("Error string should contain 'Test error', got '%s'", errorStr)
	}
}

func TestStatusWithEmptyMessage(t *testing.T) {
	status := New(INTERNAL_ERROR, "")
	
	expectedStr := "17 (INTERNAL_ERROR)"
	if status.String() != expectedStr {
		t.Fatalf("Expected '%s', got '%s'", expectedStr, status.String())
	}
}

func TestStatusThreadSafety(t *testing.T) {
	status := StatusOK
	
	// Test concurrent updates
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(i int) {
			errorStatus := New(Code(i+1), "Concurrent error")
			status.Update(errorStatus)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Status should have been updated with one of the errors
	if status.OK() {
		t.Fatal("Status should have been updated")
	}
	
	// Verify we can read the status safely
	_ = status.ErrorCode()
	_ = status.ErrorMessage()
	_ = status.String()
}