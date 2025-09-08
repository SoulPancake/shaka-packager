// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package version provides version information for the shaka packager.
package version

import (
	"fmt"
	"runtime"
)

const (
	// Major version number
	VersionMajor = 2
	// Minor version number
	VersionMinor = 6
	// Patch version number
	VersionPatch = 1
	// Version suffix (e.g., "-rc1", "-beta", "")
	VersionSuffix = ""
)

// BuildInfo contains build information.
type BuildInfo struct {
	Version     string
	GitRevision string
	BuildDate   string
	BuildType   string
	GoVersion   string
	Platform    string
}

var (
	// These will be set at build time via ldflags
	gitRevision = "unknown"
	buildDate   = "unknown"
	buildType   = "release"
)

// GetVersion returns the version string.
func GetVersion() string {
	version := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	if VersionSuffix != "" {
		version += VersionSuffix
	}
	return version
}

// GetVersionWithRevision returns the version string with git revision.
func GetVersionWithRevision() string {
	version := GetVersion()
	if gitRevision != "unknown" && gitRevision != "" {
		version += fmt.Sprintf(" (%s)", gitRevision)
	}
	return version
}

// GetFullVersion returns the full version string with all details.
func GetFullVersion() string {
	return fmt.Sprintf("Shaka Packager %s", GetVersionWithRevision())
}

// GetBuildInfo returns detailed build information.
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:     GetVersion(),
		GitRevision: gitRevision,
		BuildDate:   buildDate,
		BuildType:   buildType,
		GoVersion:   runtime.Version(),
		Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetBuildInfoString returns build information as a formatted string.
func GetBuildInfoString() string {
	info := GetBuildInfo()
	return fmt.Sprintf(
		"Version: %s\nGit Revision: %s\nBuild Date: %s\nBuild Type: %s\nGo Version: %s\nPlatform: %s",
		info.Version,
		info.GitRevision,
		info.BuildDate,
		info.BuildType,
		info.GoVersion,
		info.Platform,
	)
}

// IsReleaseBuild returns true if this is a release build.
func IsReleaseBuild() bool {
	return buildType == "release"
}

// IsDebugBuild returns true if this is a debug build.
func IsDebugBuild() bool {
	return buildType == "debug"
}