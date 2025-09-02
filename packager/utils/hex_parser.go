// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package utils

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// HexParser provides utilities for parsing hexadecimal strings.
type HexParser struct{}

// ParseHex parses a hexadecimal string and returns the bytes.
// The string can contain spaces, colons, or hyphens as separators.
func (p *HexParser) ParseHex(hexStr string) ([]byte, error) {
	// Clean the hex string by removing common separators
	cleaned := strings.ReplaceAll(hexStr, " ", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\t", "")

	// Remove 0x prefix if present
	if strings.HasPrefix(cleaned, "0x") || strings.HasPrefix(cleaned, "0X") {
		cleaned = cleaned[2:]
	}

	// Ensure even length
	if len(cleaned)%2 != 0 {
		return nil, fmt.Errorf("hex string must have even length, got: %d", len(cleaned))
	}

	// Decode hex string
	result, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %v", err)
	}

	return result, nil
}

// ToHexString converts bytes to a hexadecimal string.
func (p *HexParser) ToHexString(data []byte) string {
	return hex.EncodeToString(data)
}

// ToHexStringWithSeparator converts bytes to a hexadecimal string with separators.
func (p *HexParser) ToHexStringWithSeparator(data []byte, separator string) string {
	if len(data) == 0 {
		return ""
	}

	hexStr := hex.EncodeToString(data)
	if separator == "" {
		return hexStr
	}

	// Insert separator between each byte
	var result strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			result.WriteString(separator)
		}
		result.WriteString(hexStr[i : i+2])
	}

	return result.String()
}

// ParseKeyID parses a key ID string which can be in various formats.
func (p *HexParser) ParseKeyID(keyIDStr string) ([]byte, error) {
	return p.ParseHex(keyIDStr)
}

// ParseKey parses a key string which can be in various formats.
func (p *HexParser) ParseKey(keyStr string) ([]byte, error) {
	return p.ParseHex(keyStr)
}

// ValidateHexString validates that a string contains only valid hex characters.
func (p *HexParser) ValidateHexString(hexStr string) error {
	cleaned := strings.ReplaceAll(hexStr, " ", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	if strings.HasPrefix(cleaned, "0x") || strings.HasPrefix(cleaned, "0X") {
		cleaned = cleaned[2:]
	}

	if len(cleaned)%2 != 0 {
		return fmt.Errorf("hex string must have even length")
	}

	for _, r := range cleaned {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return fmt.Errorf("invalid hex character: %c", r)
		}
	}

	return nil
}

// Global hex parser instance
var GlobalHexParser = &HexParser{}