// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package app

import (
	"testing"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

func TestParseStreamDescriptor(t *testing.T) {
	testCases := []struct {
		input       string
		expectError bool
		expected    *packager.StreamDescriptor
	}{
		{
			input:       "input=test.mp4,output=output.mp4",
			expectError: false,
			expected: &packager.StreamDescriptor{
				Input:  "test.mp4",
				Output: "output.mp4",
			},
		},
		{
			input:       "input=test.mp4,stream=video,output=video.mp4,bandwidth=1000000",
			expectError: false,
			expected: &packager.StreamDescriptor{
				Input:     "test.mp4",
				Stream:    "video",
				Output:    "video.mp4",
				Bandwidth: 1000000,
			},
		},
		{
			input:       "in=test.mp4,out=output.mp4,lang=en",
			expectError: false,
			expected: &packager.StreamDescriptor{
				Input:    "test.mp4",
				Output:   "output.mp4",
				Language: "en",
			},
		},
		{
			input:       "input=test.mp4,segment_template=segment_$Number$.m4s",
			expectError: false,
			expected: &packager.StreamDescriptor{
				Input:           "test.mp4",
				SegmentTemplate: "segment_$Number$.m4s",
			},
		},
		{
			input:       "input=test.mp4,output=output.mp4,hls_name=audio,hls_group_id=audio-group",
			expectError: false,
			expected: &packager.StreamDescriptor{
				Input:      "test.mp4",
				Output:     "output.mp4",
				HlsName:    "audio",
				HlsGroupId: "audio-group",
			},
		},
		{
			input:       "input=test.mp4,output=output.mp4,dash_roles=main;alternate",
			expectError: false,
			expected: &packager.StreamDescriptor{
				Input:     "test.mp4",
				Output:    "output.mp4",
				DashRoles: []string{"main", "alternate"},
			},
		},
		{
			input:       "",
			expectError: true,
			expected:    nil,
		},
		{
			input:       "output=output.mp4", // Missing input
			expectError: true,
			expected:    nil,
		},
		{
			input:       "input=test.mp4", // Missing output or segment_template
			expectError: true,
			expected:    nil,
		},
		{
			input:       "input=test.mp4,bandwidth=invalid",
			expectError: true,
			expected:    nil,
		},
	}

	for i, tc := range testCases {
		result, err := ParseStreamDescriptor(tc.input)
		
		if tc.expectError {
			if err == nil {
				t.Errorf("Test case %d: expected error but got none", i)
			}
			continue
		}
		
		if err != nil {
			t.Errorf("Test case %d: unexpected error: %v", i, err)
			continue
		}
		
		if result == nil {
			t.Errorf("Test case %d: got nil result", i)
			continue
		}
		
		// Check individual fields
		if result.Input != tc.expected.Input {
			t.Errorf("Test case %d: Input mismatch. Expected: %s, Got: %s", i, tc.expected.Input, result.Input)
		}
		
		if result.Output != tc.expected.Output {
			t.Errorf("Test case %d: Output mismatch. Expected: %s, Got: %s", i, tc.expected.Output, result.Output)
		}
		
		if result.Stream != tc.expected.Stream {
			t.Errorf("Test case %d: Stream mismatch. Expected: %s, Got: %s", i, tc.expected.Stream, result.Stream)
		}
		
		if result.Bandwidth != tc.expected.Bandwidth {
			t.Errorf("Test case %d: Bandwidth mismatch. Expected: %d, Got: %d", i, tc.expected.Bandwidth, result.Bandwidth)
		}
		
		if result.Language != tc.expected.Language {
			t.Errorf("Test case %d: Language mismatch. Expected: %s, Got: %s", i, tc.expected.Language, result.Language)
		}
		
		if result.SegmentTemplate != tc.expected.SegmentTemplate {
			t.Errorf("Test case %d: SegmentTemplate mismatch. Expected: %s, Got: %s", i, tc.expected.SegmentTemplate, result.SegmentTemplate)
		}
		
		if result.HlsName != tc.expected.HlsName {
			t.Errorf("Test case %d: HlsName mismatch. Expected: %s, Got: %s", i, tc.expected.HlsName, result.HlsName)
		}
		
		if result.HlsGroupId != tc.expected.HlsGroupId {
			t.Errorf("Test case %d: HlsGroupId mismatch. Expected: %s, Got: %s", i, tc.expected.HlsGroupId, result.HlsGroupId)
		}
	}
}

func TestGetFieldType(t *testing.T) {
	testCases := []struct {
		fieldName string
		expected  FieldType
	}{
		{"input", InputField},
		{"in", InputField},
		{"output", OutputField},
		{"out", OutputField},
		{"stream", StreamSelectorField},
		{"stream_selector", StreamSelectorField},
		{"bandwidth", BandwidthField},
		{"bw", BandwidthField},
		{"bitrate", BandwidthField},
		{"language", LanguageField},
		{"lang", LanguageField},
		{"unknown_field", UnknownField},
	}
	
	for _, tc := range testCases {
		result := GetFieldType(tc.fieldName)
		if result != tc.expected {
			t.Errorf("GetFieldType(%s) = %v, expected %v", tc.fieldName, result, tc.expected)
		}
	}
}

func TestParseStreamDescriptors(t *testing.T) {
	descriptorStrings := []string{
		"input=video.mp4,stream=video,output=video_out.mp4",
		"input=audio.mp4,stream=audio,output=audio_out.mp4",
	}
	
	descriptors, err := ParseStreamDescriptors(descriptorStrings)
	if err != nil {
		t.Fatalf("ParseStreamDescriptors failed: %v", err)
	}
	
	if len(descriptors) != 2 {
		t.Errorf("Expected 2 descriptors, got %d", len(descriptors))
	}
	
	// Check first descriptor
	if descriptors[0].Input != "video.mp4" {
		t.Errorf("First descriptor input mismatch. Expected: video.mp4, Got: %s", descriptors[0].Input)
	}
	
	// Check second descriptor
	if descriptors[1].Stream != "audio" {
		t.Errorf("Second descriptor stream mismatch. Expected: audio, Got: %s", descriptors[1].Stream)
	}
}

func TestValidateDescriptor(t *testing.T) {
	testCases := []struct {
		descriptor  *packager.StreamDescriptor
		expectError bool
	}{
		{
			descriptor: &packager.StreamDescriptor{
				Input:  "test.mp4",
				Output: "output.mp4",
			},
			expectError: false,
		},
		{
			descriptor: &packager.StreamDescriptor{
				Input:           "test.mp4",
				SegmentTemplate: "segment_$Number$.m4s",
			},
			expectError: false,
		},
		{
			descriptor: &packager.StreamDescriptor{
				Output: "output.mp4", // Missing input
			},
			expectError: true,
		},
		{
			descriptor: &packager.StreamDescriptor{
				Input: "test.mp4", // Missing output and segment_template
			},
			expectError: true,
		},
	}
	
	for i, tc := range testCases {
		err := validateDescriptor(tc.descriptor)
		
		if tc.expectError && err == nil {
			t.Errorf("Test case %d: expected error but got none", i)
		} else if !tc.expectError && err != nil {
			t.Errorf("Test case %d: unexpected error: %v", i, err)
		}
	}
}