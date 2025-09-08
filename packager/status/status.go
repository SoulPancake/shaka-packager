// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package status provides status handling for the shaka packager.
package status

import (
	"fmt"
	"sync"
)

// Code represents error codes for the packager APIs.
type Code int

const (
	// Not an error; returned on success
	OK Code = iota

	// Unknown error
	UNKNOWN

	// The operation was cancelled (typically by the caller)
	CANCELLED

	// Client specified an invalid argument
	INVALID_ARGUMENT

	// Operation is not implemented or not supported/enabled
	UNIMPLEMENTED

	// Cannot open file
	FILE_FAILURE

	// End of stream
	END_OF_STREAM

	// Failure to get HTTP response successfully
	HTTP_FAILURE

	// Unable to parse the media file
	PARSER_FAILURE

	// Failed to do the encryption
	ENCRYPTION_FAILURE

	// Error when trying to do chunking
	CHUNKING_ERROR

	// Fail to mux the media file
	MUXER_FAILURE

	// This track fragment is finalized
	FRAGMENT_FINALIZED

	// Server errors. Receives malformed response from server
	SERVER_ERROR

	// Internal errors. Some invariants have been broken
	INTERNAL_ERROR

	// The operation was stopped
	STOPPED

	// The operation timed out
	TIME_OUT

	// Value was not found
	NOT_FOUND

	// The entity that a client attempted to create already exists
	ALREADY_EXISTS

	// Error when trying to generate trick play stream
	TRICK_PLAY_ERROR
)

var codeNames = map[Code]string{
	OK:                   "OK",
	UNKNOWN:              "UNKNOWN",
	CANCELLED:            "CANCELLED",
	INVALID_ARGUMENT:     "INVALID_ARGUMENT",
	UNIMPLEMENTED:        "UNIMPLEMENTED",
	FILE_FAILURE:         "FILE_FAILURE",
	END_OF_STREAM:        "END_OF_STREAM",
	HTTP_FAILURE:         "HTTP_FAILURE",
	PARSER_FAILURE:       "PARSER_FAILURE",
	ENCRYPTION_FAILURE:   "ENCRYPTION_FAILURE",
	CHUNKING_ERROR:       "CHUNKING_ERROR",
	MUXER_FAILURE:        "MUXER_FAILURE",
	FRAGMENT_FINALIZED:   "FRAGMENT_FINALIZED",
	SERVER_ERROR:         "SERVER_ERROR",
	INTERNAL_ERROR:       "INTERNAL_ERROR",
	STOPPED:              "STOPPED",
	TIME_OUT:             "TIME_OUT",
	NOT_FOUND:            "NOT_FOUND",
	ALREADY_EXISTS:       "ALREADY_EXISTS",
	TRICK_PLAY_ERROR:     "TRICK_PLAY_ERROR",
}

// String returns the string representation of the error code.
func (c Code) String() string {
	if name, ok := codeNames[c]; ok {
		return name
	}
	return "UNKNOWN_STATUS"
}

// Status represents the status of an operation.
type Status struct {
	mu           sync.RWMutex
	errorCode    Code
	errorMessage string
}

var (
	// Pre-defined Status objects
	StatusOK      = Status{errorCode: OK}
	StatusUnknown = Status{errorCode: UNKNOWN}
)

// New creates a status with the specified code and error message.
// If errorCode == OK, error message is ignored and a Status
// object identical to StatusOK is constructed.
func New(errorCode Code, errorMessage string) Status {
	if errorCode == OK {
		return StatusOK
	}
	return Status{
		errorCode:    errorCode,
		errorMessage: errorMessage,
	}
}

// Update stores newStatus into this status if this status is OK.
// If this status is not OK, preserves the current error code and message.
// This is a convenient way of keeping track of the first error encountered.
func (s *Status) Update(newStatus Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.errorCode == OK {
		s.errorCode = newStatus.errorCode
		s.errorMessage = newStatus.errorMessage
	}
}

// OK returns true if the status represents success.
func (s *Status) OK() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errorCode == OK
}

// ErrorCode returns the error code.
func (s *Status) ErrorCode() Code {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errorCode
}

// ErrorMessage returns the error message.
func (s *Status) ErrorMessage() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errorMessage
}

// Equal checks if two Status objects are equal.
func (s *Status) Equal(other Status) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	
	return s.errorCode == other.errorCode && s.errorMessage == other.errorMessage
}

// String returns a combination of the error code name and message.
func (s *Status) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.errorCode == OK {
		return "OK"
	}
	
	if s.errorMessage == "" {
		return fmt.Sprintf("%d (%s)", int(s.errorCode), s.errorCode.String())
	}
	
	return fmt.Sprintf("%d (%s): %s", int(s.errorCode), s.errorCode.String(), s.errorMessage)
}

// Error implements the error interface, allowing Status to be used as an error.
func (s *Status) Error() string {
	if s.OK() {
		return ""
	}
	return s.String()
}