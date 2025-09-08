// Copyright 2016 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package hls

import (
	"strings"
	"testing"
	"time"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

func TestMediaPlaylistBasic(t *testing.T) {
	hlsParams := packager.HlsParams{
		TargetDuration: 10,
	}
	
	playlist := NewMediaPlaylist(hlsParams, "playlist.m3u8", "audio", "audio-group")
	
	if playlist.FileName() != "playlist.m3u8" {
		t.Errorf("Expected filename 'playlist.m3u8', got '%s'", playlist.FileName())
	}
	
	if playlist.Name() != "audio" {
		t.Errorf("Expected name 'audio', got '%s'", playlist.Name())
	}
	
	if playlist.GroupId() != "audio-group" {
		t.Errorf("Expected group ID 'audio-group', got '%s'", playlist.GroupId())
	}
}

func TestMediaPlaylistAddSegment(t *testing.T) {
	hlsParams := packager.HlsParams{
		TargetDuration: 10,
	}
	
	playlist := NewMediaPlaylist(hlsParams, "playlist.m3u8", "video", "video-group")
	playlist.SetStreamType(StreamTypeVideo)
	
	// Add segments
	playlist.AddSegment("segment001.m4s", 9.5, "")
	playlist.AddSegment("segment002.m4s", 10.0, "")
	playlist.AddSegment("segment003.m4s", 8.5, "")
	
	output := playlist.ToString()
	
	// Check basic structure
	if !strings.Contains(output, "#EXTM3U") {
		t.Error("Playlist should contain #EXTM3U header")
	}
	
	if !strings.Contains(output, "#EXT-X-VERSION:6") {
		t.Error("Playlist should contain version 6")
	}
	
	if !strings.Contains(output, "#EXT-X-TARGETDURATION:10") {
		t.Error("Playlist should contain target duration")
	}
	
	if !strings.Contains(output, "segment001.m4s") {
		t.Error("Playlist should contain segment001.m4s")
	}
	
	if !strings.Contains(output, "#EXTINF:9.500000") {
		t.Error("Playlist should contain correct duration for first segment")
	}
	
	if !strings.Contains(output, "#EXT-X-ENDLIST") {
		t.Error("VOD playlist should contain endlist")
	}
}

func TestMediaPlaylistEncryption(t *testing.T) {
	hlsParams := packager.HlsParams{
		TargetDuration: 10,
	}
	
	playlist := NewMediaPlaylist(hlsParams, "playlist.m3u8", "video", "video-group")
	
	// Add encryption info
	playlist.AddEncryptionInfo(EncryptionAes128, "https://example.com/key.bin", "0x12345678901234567890123456789012", "", "")
	
	// Add segment
	playlist.AddSegment("segment001.m4s", 10.0, "")
	
	output := playlist.ToString()
	
	if !strings.Contains(output, "#EXT-X-KEY:METHOD=AES-128") {
		t.Error("Playlist should contain encryption method")
	}
	
	if !strings.Contains(output, "URI=\"https://example.com/key.bin\"") {
		t.Error("Playlist should contain key URI")
	}
	
	if !strings.Contains(output, "IV=0x12345678901234567890123456789012") {
		t.Error("Playlist should contain IV")
	}
}

func TestMediaPlaylistLive(t *testing.T) {
	hlsParams := packager.HlsParams{
		TargetDuration: 6,
	}
	
	playlist := NewMediaPlaylist(hlsParams, "live.m3u8", "live-video", "video-group")
	playlist.SetIsLive(true)
	playlist.SetSequenceNumber(100)
	
	// Add segments
	playlist.AddSegment("seg100.m4s", 6.0, "")
	playlist.AddSegment("seg101.m4s", 6.0, "")
	
	output := playlist.ToString()
	
	if !strings.Contains(output, "#EXT-X-MEDIA-SEQUENCE:100") {
		t.Error("Live playlist should contain media sequence")
	}
	
	if strings.Contains(output, "#EXT-X-PLAYLIST-TYPE:") {
		t.Error("Live playlist should not contain playlist type")
	}
	
	if strings.Contains(output, "#EXT-X-ENDLIST") {
		t.Error("Live playlist should not contain endlist")
	}
}

func TestMediaPlaylistDiscontinuity(t *testing.T) {
	hlsParams := packager.HlsParams{
		TargetDuration: 10,
	}
	
	playlist := NewMediaPlaylist(hlsParams, "playlist.m3u8", "video", "video-group")
	
	// Add segments with discontinuity
	playlist.AddSegment("segment001.m4s", 10.0, "")
	playlist.AddDiscontinuity()
	playlist.AddSegment("segment002.m4s", 10.0, "")
	
	output := playlist.ToString()
	
	if !strings.Contains(output, "#EXT-X-DISCONTINUITY") {
		t.Error("Playlist should contain discontinuity marker")
	}
	
	// Check order
	lines := strings.Split(output, "\n")
	foundDiscontinuity := false
	foundSecondSegment := false
	
	for _, line := range lines {
		if line == "#EXT-X-DISCONTINUITY" {
			foundDiscontinuity = true
		}
		if foundDiscontinuity && strings.Contains(line, "segment002.m4s") {
			foundSecondSegment = true
			break
		}
	}
	
	if !foundSecondSegment {
		t.Error("Discontinuity should appear before second segment")
	}
}

func TestExtInfEntry(t *testing.T) {
	entry := &ExtInfEntry{
		Duration: 9.5,
		Title:    "Test Title",
		Uri:      "segment.m4s",
	}
	
	output := entry.ToString()
	expected := "#EXTINF:9.500000,Test Title\nsegment.m4s"
	
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestExtKeyEntry(t *testing.T) {
	entry := &ExtKeyEntry{
		Method:            EncryptionAes128,
		Uri:               "https://example.com/key.bin",
		IV:                "0x12345678901234567890123456789012",
		KeyFormat:         "identity",
		KeyFormatVersions: "1",
	}
	
	output := entry.ToString()
	
	expected := "#EXT-X-KEY:METHOD=AES-128,URI=\"https://example.com/key.bin\",IV=0x12345678901234567890123456789012,KEYFORMAT=\"identity\",KEYFORMATVERSIONS=\"1\""
	
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestEncryptionMethodString(t *testing.T) {
	testCases := []struct {
		method   EncryptionMethod
		expected string
	}{
		{EncryptionNone, "NONE"},
		{EncryptionAes128, "AES-128"},
		{EncryptionSampleAes, "SAMPLE-AES"},
		{EncryptionSampleAesCenc, "SAMPLE-AES-CENC"},
	}
	
	for _, tc := range testCases {
		result := tc.method.String()
		if result != tc.expected {
			t.Errorf("EncryptionMethod(%d).String() = %s, expected %s", tc.method, result, tc.expected)
		}
	}
}

func TestStreamTypeString(t *testing.T) {
	testCases := []struct {
		streamType MediaPlaylistStreamType
		expected   string
	}{
		{StreamTypeAudio, "AUDIO"},
		{StreamTypeVideo, "VIDEO"},
		{StreamTypeVideoIFramesOnly, "VIDEO"},
		{StreamTypeSubtitle, "SUBTITLES"},
		{StreamTypeUnknown, "UNKNOWN"},
	}
	
	for _, tc := range testCases {
		result := tc.streamType.String()
		if result != tc.expected {
			t.Errorf("StreamType(%d).String() = %s, expected %s", tc.streamType, result, tc.expected)
		}
	}
}

func TestMediaPlaylistTargetDuration(t *testing.T) {
	hlsParams := packager.HlsParams{
		TargetDuration: 15,
	}
	
	playlist := NewMediaPlaylist(hlsParams, "playlist.m3u8", "test", "test-group")
	
	// Override target duration
	playlist.SetTargetDuration(12 * time.Second)
	
	// Add a segment longer than initial target duration but within new one
	playlist.AddSegment("segment001.m4s", 11.0, "")
	
	output := playlist.ToString()
	
	if !strings.Contains(output, "#EXT-X-TARGETDURATION:12") {
		t.Error("Playlist should contain updated target duration of 12")
	}
}