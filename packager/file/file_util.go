// Copyright 2016 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TempFilePath creates a temp file name in directory tempDir
// Generate the temp file in OS specific temporary directory if tempDir is empty
// Returns the temp file path on success, error otherwise
func TempFilePath(tempDir string) (string, error) {
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	
	// Create a unique filename using timestamp and random suffix
	timestamp := time.Now().UnixNano()
	tempFileName := fmt.Sprintf("shaka_temp_%d", timestamp)
	
	tempFilePath := filepath.Join(tempDir, tempFileName)
	
	// Ensure the directory exists
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	
	return tempFilePath, nil
}

// MakePathRelative makes a media path relative to a parent path
func MakePathRelative(mediaPath, parentPath string) string {
	// Clean the paths to handle any . or .. components
	cleanMediaPath := filepath.Clean(mediaPath)
	cleanParentPath := filepath.Clean(parentPath)
	
	// Convert to absolute paths for comparison
	absMediaPath, err := filepath.Abs(cleanMediaPath)
	if err != nil {
		return cleanMediaPath // Return original if we can't make it absolute
	}
	
	absParentPath, err := filepath.Abs(cleanParentPath)
	if err != nil {
		return cleanMediaPath // Return original if we can't make parent absolute
	}
	
	// Try to get relative path
	relPath, err := filepath.Rel(absParentPath, absMediaPath)
	if err != nil {
		return cleanMediaPath // Return original if we can't make it relative
	}
	
	// Convert back slashes to forward slashes for consistency across platforms
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	
	return relPath
}

// FileExists checks if a file exists and is not a directory
func FileExists(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists checks if a directory exists
func DirectoryExists(dirpath string) bool {
	info, err := os.Stat(dirpath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CreateDirectories creates directories recursively if they don't exist
func CreateDirectories(dirpath string) error {
	return os.MkdirAll(dirpath, 0755)
}

// GetFileSize returns the size of a file
func GetFileSize(filepath string) (int64, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// DeleteFile deletes a file
func DeleteFile(filepath string) error {
	return os.Remove(filepath)
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	buffer := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := sourceFile.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			return err
		}
		if n == 0 {
			break
		}
		
		_, err = destFile.Write(buffer[:n])
		if err != nil {
			return err
		}
	}
	
	return destFile.Sync()
}