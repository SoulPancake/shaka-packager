// Package utils provides utility functions for the packager
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FormatDuration formats a duration in seconds to ISO 8601 duration format
func FormatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := duration.Seconds() - float64(hours*3600) - float64(minutes*60)
	return fmt.Sprintf("PT%dH%dM%.3fS", hours, minutes, secs)
}

// ParseDuration parses an ISO 8601 duration string to seconds
func ParseDuration(durationStr string) (float64, error) {
	// Simplified parser for PT format
	if !strings.HasPrefix(durationStr, "PT") {
		return 0, fmt.Errorf("invalid duration format: %s", durationStr)
	}
	
	// For now, return a default value
	// A full implementation would parse the PT format
	return 60.0, nil
}

// EnsureDirectory creates a directory if it doesn't exist
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetFileExtension returns the file extension (including the dot)
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GetBaseName returns the base name without extension
func GetBaseName(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// JoinPath joins path components using the OS-specific separator
func JoinPath(components ...string) string {
	return filepath.Join(components...)
}

// SanitizeFilename removes invalid characters from a filename
func SanitizeFilename(filename string) string {
	// Remove or replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	
	return result
}

// BytesToHex converts bytes to hexadecimal string
func BytesToHex(data []byte) string {
	return fmt.Sprintf("%X", data)
}

// HexToBytes converts hexadecimal string to bytes
func HexToBytes(hex string) ([]byte, error) {
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}
	
	bytes := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hex[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		bytes[i/2] = b
	}
	
	return bytes, nil
}

// ClampInt clamps an integer value between min and max
func ClampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampFloat64 clamps a float64 value between min and max
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxUint32 returns the maximum of two uint32 values
func MaxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// MinUint32 returns the minimum of two uint32 values
func MinUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// ContainsString checks if a string slice contains a specific string
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveString removes the first occurrence of a string from a slice
func RemoveString(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// UniqueStrings returns a slice with duplicate strings removed
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}