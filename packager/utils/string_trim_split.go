// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package utils

import (
	"strings"
)

// StringTrimSplit provides utilities for string trimming and splitting.
type StringTrimSplit struct{}

// TrimAndSplit trims whitespace from a string and splits it by the given separator.
func (s *StringTrimSplit) TrimAndSplit(input, separator string) []string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return []string{}
	}
	
	parts := strings.Split(trimmed, separator)
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart != "" {
			result = append(result, trimmedPart)
		}
	}
	
	return result
}

// SplitAndTrim splits a string by the given separator and trims each part.
func (s *StringTrimSplit) SplitAndTrim(input, separator string) []string {
	if input == "" {
		return []string{}
	}
	
	parts := strings.Split(input, separator)
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart != "" {
			result = append(result, trimmedPart)
		}
	}
	
	return result
}

// TrimLines trims whitespace from each line in a multi-line string.
func (s *StringTrimSplit) TrimLines(input string) []string {
	lines := strings.Split(input, "\n")
	result := make([]string, 0, len(lines))
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		result = append(result, trimmedLine)
	}
	
	return result
}

// TrimAndSplitLines trims whitespace and splits by lines, removing empty lines.
func (s *StringTrimSplit) TrimAndSplitLines(input string) []string {
	lines := strings.Split(input, "\n")
	result := make([]string, 0, len(lines))
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			result = append(result, trimmedLine)
		}
	}
	
	return result
}

// SplitKeyValue splits a key=value string into key and value parts.
func (s *StringTrimSplit) SplitKeyValue(input, separator string) (string, string, bool) {
	parts := strings.SplitN(input, separator, 2)
	if len(parts) != 2 {
		return "", "", false
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	
	return key, value, true
}

// ParseCommaSeparatedValues parses a comma-separated string into a slice.
func (s *StringTrimSplit) ParseCommaSeparatedValues(input string) []string {
	return s.TrimAndSplit(input, ",")
}

// ParseSemicolonSeparatedValues parses a semicolon-separated string into a slice.
func (s *StringTrimSplit) ParseSemicolonSeparatedValues(input string) []string {
	return s.TrimAndSplit(input, ";")
}

// JoinWithCommas joins string slice with commas.
func (s *StringTrimSplit) JoinWithCommas(values []string) string {
	return strings.Join(values, ", ")
}

// Global string trim split instance
var GlobalStringTrimSplit = &StringTrimSplit{}