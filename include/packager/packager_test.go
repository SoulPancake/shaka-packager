// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

import (
	"testing"
	"time"
)

func TestPackagingParamsDefaults(t *testing.T) {
	params := NewPackagingParams()
	
	// Test default values
	if !params.Mp4OutputParams.IncludePSSHInStream {
		t.Fatal("IncludePSSHInStream should default to true")
	}
	
	if !params.Mp4OutputParams.GenerateSIDXInMediaSegments {
		t.Fatal("GenerateSIDXInMediaSegments should default to true")
	}
	
	if params.ChunkingParams.StartSegmentNumber != 1 {
		t.Fatal("StartSegmentNumber should default to 1")
	}
	
	if params.EncryptionParams.KeyProvider != KeyProviderNone {
		t.Fatal("KeyProvider should default to None")
	}
}

func TestStreamDescriptor(t *testing.T) {
	desc := NewStreamDescriptor()
	
	if desc.CcIndex != -1 {
		t.Fatal("CcIndex should default to -1")
	}
	
	// Test setting values
	desc.Input = "input.mp4"
	desc.StreamSelector = "video"
	desc.Output = "output.mp4"
	
	if desc.Input != "input.mp4" {
		t.Fatalf("Expected 'input.mp4', got '%s'", desc.Input)
	}
}

func TestPackagerLifecycle(t *testing.T) {
	packager := NewPackager()
	defer packager.Close()
	
	// Test initialization with invalid parameters
	params := NewPackagingParams()
	descriptors := []StreamDescriptor{} // Empty descriptors should fail
	
	status := packager.Initialize(params, descriptors)
	if status.OK() {
		t.Fatal("Initialize should fail with empty descriptors")
	}
	
	if status.ErrorCode() != INVALID_ARGUMENT {
		t.Fatalf("Expected INVALID_ARGUMENT, got %d", status.ErrorCode())
	}
	
	// Test with valid parameters
	descriptors = []StreamDescriptor{
		{
			Input:          "input.mp4",
			StreamSelector: "video",
			Output:         "output.mp4",
		},
	}
	
	status = packager.Initialize(params, descriptors)
	if !status.OK() {
		t.Fatalf("Initialize should succeed: %s", status.String())
	}
	
	// Test double initialization
	status = packager.Initialize(params, descriptors)
	if status.OK() {
		t.Fatal("Double initialization should fail")
	}
}

func TestEncryptionParams(t *testing.T) {
	params := NewEncryptionParams()
	
	// Test defaults
	if params.KeyProvider != KeyProviderNone {
		t.Fatal("KeyProvider should default to None")
	}
	
	if params.ProtectionScheme != ProtectionSchemeCENC {
		t.Fatal("ProtectionScheme should default to CENC")
	}
	
	if params.CryptByteBlock != DefaultCryptByteBlock {
		t.Fatalf("CryptByteBlock should default to %d", DefaultCryptByteBlock)
	}
	
	if params.SkipByteBlock != DefaultSkipByteBlock {
		t.Fatalf("SkipByteBlock should default to %d", DefaultSkipByteBlock)
	}
}

func TestHlsParams(t *testing.T) {
	params := NewHlsParams()
	
	if params.PlaylistType != HlsPlaylistTypeVod {
		t.Fatal("PlaylistType should default to VOD")
	}
	
	if params.StartTimeOffset != nil {
		t.Fatal("StartTimeOffset should default to nil")
	}
	
	// Test setting optional values
	offset := 10.5
	params.StartTimeOffset = &offset
	
	if *params.StartTimeOffset != 10.5 {
		t.Fatalf("Expected 10.5, got %f", *params.StartTimeOffset)
	}
}

func TestMpdParams(t *testing.T) {
	params := NewMpdParams()
	
	if params.MinBufferTime != 2.0 {
		t.Fatalf("MinBufferTime should default to 2.0, got %f", params.MinBufferTime)
	}
	
	if !params.UseSegmentTimeline {
		t.Fatal("UseSegmentTimeline should default to true")
	}
	
	// Test duration field
	params.DefaultPresentationDuration = time.Duration(30) * time.Second
	
	if params.DefaultPresentationDuration != 30*time.Second {
		t.Fatal("DefaultPresentationDuration not set correctly")
	}
}

func TestDefaultStreamLabelFunction(t *testing.T) {
	// Test audio stream
	audioAttrs := EncryptedStreamAttributes{
		StreamType: StreamTypeAudio,
	}
	
	label := DefaultStreamLabelFunction(720*480, 1280*720, 1920*1080, audioAttrs)
	if label != "AUDIO" {
		t.Fatalf("Expected 'AUDIO', got '%s'", label)
	}
	
	// Test SD video
	sdVideoAttrs := EncryptedStreamAttributes{
		StreamType: StreamTypeVideo,
	}
	sdVideoAttrs.Video.Width = 640
	sdVideoAttrs.Video.Height = 480
	
	label = DefaultStreamLabelFunction(720*480, 1280*720, 1920*1080, sdVideoAttrs)
	if label != "SD" {
		t.Fatalf("Expected 'SD', got '%s'", label)
	}
	
	// Test HD video
	hdVideoAttrs := EncryptedStreamAttributes{
		StreamType: StreamTypeVideo,
	}
	hdVideoAttrs.Video.Width = 1280
	hdVideoAttrs.Video.Height = 720
	
	label = DefaultStreamLabelFunction(720*480, 1280*720, 1920*1080, hdVideoAttrs)
	if label != "HD" {
		t.Fatalf("Expected 'HD', got '%s'", label)
	}
	
	// Test UHD1 video
	uhd1VideoAttrs := EncryptedStreamAttributes{
		StreamType: StreamTypeVideo,
	}
	uhd1VideoAttrs.Video.Width = 1920
	uhd1VideoAttrs.Video.Height = 1080
	
	label = DefaultStreamLabelFunction(720*480, 1280*720, 1920*1080, uhd1VideoAttrs)
	if label != "UHD1" {
		t.Fatalf("Expected 'UHD1', got '%s'", label)
	}
	
	// Test UHD2 video
	uhd2VideoAttrs := EncryptedStreamAttributes{
		StreamType: StreamTypeVideo,
	}
	uhd2VideoAttrs.Video.Width = 3840
	uhd2VideoAttrs.Video.Height = 2160
	
	label = DefaultStreamLabelFunction(720*480, 1280*720, 1920*1080, uhd2VideoAttrs)
	if label != "UHD2" {
		t.Fatalf("Expected 'UHD2', got '%s'", label)
	}
}

func TestEnumStringMethods(t *testing.T) {
	// Test KeyProvider
	if KeyProviderRawKey.String() != "RawKey" {
		t.Fatalf("Expected 'RawKey', got '%s'", KeyProviderRawKey.String())
	}
	
	// Test ProtectionSystem
	combined := ProtectionSystemWidevine.Or(ProtectionSystemPlayReady)
	expectedStr := "Widevine|PlayReady"
	if combined.String() != expectedStr {
		t.Fatalf("Expected '%s', got '%s'", expectedStr, combined.String())
	}
	
	// Test HlsPlaylistType
	if HlsPlaylistTypeLive.String() != "LIVE" {
		t.Fatalf("Expected 'LIVE', got '%s'", HlsPlaylistTypeLive.String())
	}
}

func TestProtectionSystemOperations(t *testing.T) {
	// Test OR operation
	combined := ProtectionSystemWidevine.Or(ProtectionSystemPlayReady)
	
	if !combined.HasFlag(ProtectionSystemWidevine) {
		t.Fatal("Combined should have Widevine flag")
	}
	
	if !combined.HasFlag(ProtectionSystemPlayReady) {
		t.Fatal("Combined should have PlayReady flag")
	}
	
	if combined.HasFlag(ProtectionSystemFairPlay) {
		t.Fatal("Combined should not have FairPlay flag")
	}
	
	// Test AND operation
	result := combined.And(ProtectionSystemWidevine)
	if result != ProtectionSystemWidevine {
		t.Fatal("AND operation failed")
	}
}