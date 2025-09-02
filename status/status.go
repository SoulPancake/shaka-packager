// Package status provides status codes and error handling for the packager
package status

import "fmt"

// Status represents a status code with optional message
type Status struct {
	Code    Code
	Message string
}

// Code represents status codes
type Code int

const (
	OK Code = iota
	Cancelled
	Unknown
	InvalidArgument
	DeadlineExceeded
	NotFound
	AlreadyExists
	PermissionDenied
	ResourceExhausted
	FailedPrecondition
	Aborted
	OutOfRange
	Unimplemented
	Internal
	Unavailable
	DataLoss
	Unauthenticated
)

// String returns the string representation of the status code
func (c Code) String() string {
	switch c {
	case OK:
		return "OK"
	case Cancelled:
		return "CANCELLED"
	case Unknown:
		return "UNKNOWN"
	case InvalidArgument:
		return "INVALID_ARGUMENT"
	case DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case NotFound:
		return "NOT_FOUND"
	case AlreadyExists:
		return "ALREADY_EXISTS"
	case PermissionDenied:
		return "PERMISSION_DENIED"
	case ResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case FailedPrecondition:
		return "FAILED_PRECONDITION"
	case Aborted:
		return "ABORTED"
	case OutOfRange:
		return "OUT_OF_RANGE"
	case Unimplemented:
		return "UNIMPLEMENTED"
	case Internal:
		return "INTERNAL"
	case Unavailable:
		return "UNAVAILABLE"
	case DataLoss:
		return "DATA_LOSS"
	case Unauthenticated:
		return "UNAUTHENTICATED"
	default:
		return fmt.Sprintf("Code(%d)", int(c))
	}
}

// Error returns the error representation of the status
func (s Status) Error() string {
	if s.Message == "" {
		return s.Code.String()
	}
	return fmt.Sprintf("%s: %s", s.Code.String(), s.Message)
}

// NewStatus creates a new status with code and message
func NewStatus(code Code, message string) Status {
	return Status{
		Code:    code,
		Message: message,
	}
}

// NewOKStatus creates a new OK status
func NewOKStatus() Status {
	return Status{Code: OK}
}

// IsOK returns true if the status represents success
func (s Status) IsOK() bool {
	return s.Code == OK
}

// IsCancelled returns true if the status represents cancellation
func (s Status) IsCancelled() bool {
	return s.Code == Cancelled
}

// IsError returns true if the status represents an error
func (s Status) IsError() bool {
	return s.Code != OK
}