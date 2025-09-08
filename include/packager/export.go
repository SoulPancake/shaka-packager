// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// Export constants and macros for the packager library.

// Version information
const (
	// Library version
	LibraryVersion = "2.6.1"
	
	// Build information  
	BuildType    = "Release"
	BuildNumber  = "1"
	GitRevision  = "unknown"
)

// Common constants used throughout the library
const (
	// Default buffer sizes
	DefaultBufferSize = 65536
	
	// Timeout values
	DefaultTimeoutMs = 30000
	
	// Common error messages
	ErrorInvalidInput    = "Invalid input"
	ErrorNotInitialized  = "Not initialized"
	ErrorAlreadyRunning  = "Already running"
	ErrorNotRunning      = "Not running"
)

// Build information structure
type BuildInfo struct {
	Version     string
	BuildType   string
	BuildNumber string
	GitRevision string
}

// GetBuildInfo returns build information for the library
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:     LibraryVersion,
		BuildType:   BuildType,
		BuildNumber: BuildNumber,
		GitRevision: GitRevision,
	}
}