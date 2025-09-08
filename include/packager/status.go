// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package packager provides the public API for the shaka packager library.
package packager

import (
	"fmt"
)

// Error codes for the packager APIs.
type Code int

const (
	// Not an error; returned on success
	OK Code = iota

	// Unknown error. An example of where this error may be returned is
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	UNKNOWN

	// The operation was cancelled (typically by the caller).
	CANCELLED

	// Client specified an invalid argument. INVALID_ARGUMENT indicates
	// arguments that are problematic regardless of the state of the system
	// (e.g. a malformed file name).
	INVALID_ARGUMENT

	// Operation is not implemented or not supported/enabled.
	UNIMPLEMENTED

	// Cannot open file.
	FILE_FAILURE

	// End of stream.
	END_OF_STREAM

	// Failure to get HTTP response successfully,
	HTTP_FAILURE

	// Unable to parse the media file.
	PARSER_FAILURE

	// Failed to do the encryption.
	ENCRYPTION_FAILURE

	// Error when trying to do chunking.
	CHUNKING_ERROR

	// Fail to mux the media file.
	MUXER_FAILURE

	// This track fragment is finalized.
	FRAGMENT_FINALIZED

	// Server errors. Receives malformed response from server.
	SERVER_ERROR

	// Internal errors. Some invariants have been broken.
	INTERNAL_ERROR

	// The operation was stopped.
	STOPPED

	// The operation timed out.
	TIME_OUT

	// Value was not found.
	NOT_FOUND

	// The entity that a client attempted to create (e.g., file or directory)
	// already exists.
	ALREADY_EXISTS

	// Error when trying to generate trick play stream.
	TRICK_PLAY_ERROR
)

// ErrorCodeToString converts an error code to its string representation.
func ErrorCodeToString(errorCode Code) string {
	switch errorCode {
	case OK:
		return "OK"
	case UNKNOWN:
		return "UNKNOWN"
	case CANCELLED:
		return "CANCELLED"
	case INVALID_ARGUMENT:
		return "INVALID_ARGUMENT"
	case UNIMPLEMENTED:
		return "UNIMPLEMENTED"
	case FILE_FAILURE:
		return "FILE_FAILURE"
	case END_OF_STREAM:
		return "END_OF_STREAM"
	case HTTP_FAILURE:
		return "HTTP_FAILURE"
	case PARSER_FAILURE:
		return "PARSER_FAILURE"
	case ENCRYPTION_FAILURE:
		return "ENCRYPTION_FAILURE"
	case CHUNKING_ERROR:
		return "CHUNKING_ERROR"
	case MUXER_FAILURE:
		return "MUXER_FAILURE"
	case FRAGMENT_FINALIZED:
		return "FRAGMENT_FINALIZED"
	case SERVER_ERROR:
		return "SERVER_ERROR"
	case INTERNAL_ERROR:
		return "INTERNAL_ERROR"
	case STOPPED:
		return "STOPPED"
	case TIME_OUT:
		return "TIME_OUT"
	case NOT_FOUND:
		return "NOT_FOUND"
	case ALREADY_EXISTS:
		return "ALREADY_EXISTS"
	case TRICK_PLAY_ERROR:
		return "TRICK_PLAY_ERROR"
	default:
		return "UNKNOWN_STATUS"
	}
}

// Status represents the status of an operation.
type Status struct {
	errorCode    Code
	errorMessage string
}

// Pre-defined Status objects.
var (
	StatusOK      = Status{errorCode: OK}
	StatusUnknown = Status{errorCode: UNKNOWN}
)

// NewStatus creates a status with the specified code and error message.
// If "errorCode == OK", error message is ignored and a Status
// object identical to StatusOK is constructed.
func NewStatus(errorCode Code, errorMessage string) Status {
	if errorCode == OK {
		return StatusOK
	}
	return Status{
		errorCode:    errorCode,
		errorMessage: errorMessage,
	}
}

// Update stores "newStatus" into this status if this status is OK.
// If this status is not OK, preserves the current error code and message.
//
// Convenient way of keeping track of the first error encountered.
// Instead of:
//   if overallStatus.OK() { overallStatus = newStatus }
// Use:
//   overallStatus.Update(newStatus)
func (s *Status) Update(newStatus Status) {
	if s.OK() {
		*s = newStatus
	}
}

// OK returns true if the status represents success.
func (s Status) OK() bool {
	return s.errorCode == OK
}

// ErrorCode returns the error code.
func (s Status) ErrorCode() Code {
	return s.errorCode
}

// ErrorMessage returns the error message.
func (s Status) ErrorMessage() string {
	return s.errorMessage
}

// Equal checks if two Status objects are equal.
func (s Status) Equal(other Status) bool {
	return s.errorCode == other.errorCode && s.errorMessage == other.errorMessage
}

// ToString returns a combination of the error code name and message.
func (s Status) ToString() string {
	if s.errorCode == OK {
		return "OK"
	}
	
	if s.errorMessage == "" {
		return fmt.Sprintf("%d (%s)", int(s.errorCode), ErrorCodeToString(s.errorCode))
	}
	
	return fmt.Sprintf("%d (%s): %s", int(s.errorCode), ErrorCodeToString(s.errorCode), s.errorMessage)
}

// String implements the Stringer interface.
func (s Status) String() string {
	return s.ToString()
}

// Error implements the error interface, allowing Status to be used as an error.
func (s Status) Error() string {
	if s.OK() {
		return ""
	}
	return s.ToString()
}