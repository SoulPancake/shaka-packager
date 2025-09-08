// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package base

import (
	"fmt"
	"strings"
	"time"
)

// StreamType represents different types of media streams
type StreamType int

const (
	StreamUnknown StreamType = iota
	StreamAudio
	StreamVideo
	StreamText
)

// StreamTypeToString converts stream type to string representation
func StreamTypeToString(streamType StreamType) string {
	switch streamType {
	case StreamAudio:
		return "Audio"
	case StreamVideo:
		return "Video"
	case StreamText:
		return "Text"
	default:
		return "Unknown"
	}
}

// Codec represents different media codecs
type Codec int

const (
	UnknownCodec Codec = iota
	
	// Video codecs (starting from 100)
	CodecVideo      = 100
	CodecAV1        = CodecVideo
	CodecH264       = 101
	CodecH265       = 102
	CodecH265DolbyVision = 103
	CodecVP8        = 104
	CodecVP9        = 105
	CodecVideoMaxPlusOne = 199
	
	// Audio codecs (starting from 200)
	CodecAudio      = 200
	CodecAAC        = CodecAudio
	CodecAC3        = 201
	CodecAC4        = 202
	CodecALAC       = 203
	CodecDTSC       = 204
	CodecDTSE       = 205
	CodecDTSH       = 206
	CodecDTSL       = 207
	CodecDTSM       = 208
	CodecDTSP       = 209
	CodecDTSX       = 210
	CodecEAC3       = 211
	CodecFlac       = 212
	CodecIAMF       = 213
	CodecOpus       = 214
	CodecPcm        = 215
	CodecVorbis     = 216
	CodecMP3        = 217
	CodecMha1       = 218
	CodecMhm1       = 219
	CodecAudioMaxPlusOne = 299
	
	// Text codecs (starting from 300)
	CodecText       = 300
	CodecWebVtt     = CodecText
	CodecTtml       = 301
)

// CodecToString converts codec to string representation
func CodecToString(codec Codec) string {
	switch codec {
	case CodecAV1:
		return "AV1"
	case CodecH264:
		return "H.264"
	case CodecH265:
		return "H.265"
	case CodecH265DolbyVision:
		return "H.265 Dolby Vision"
	case CodecVP8:
		return "VP8"
	case CodecVP9:
		return "VP9"
	case CodecAAC:
		return "AAC"
	case CodecAC3:
		return "AC-3"
	case CodecAC4:
		return "AC-4"
	case CodecALAC:
		return "ALAC"
	case CodecDTSC:
		return "DTS-C"
	case CodecDTSE:
		return "DTS-E"
	case CodecDTSH:
		return "DTS-H"
	case CodecDTSL:
		return "DTS-L"
	case CodecDTSM:
		return "DTS-M"
	case CodecDTSP:
		return "DTS-P"
	case CodecDTSX:
		return "DTS-X"
	case CodecEAC3:
		return "E-AC-3"
	case CodecFlac:
		return "FLAC"
	case CodecIAMF:
		return "IAMF"
	case CodecOpus:
		return "Opus"
	case CodecPcm:
		return "PCM"
	case CodecVorbis:
		return "Vorbis"
	case CodecMP3:
		return "MP3"
	case CodecMha1:
		return "MHA1"
	case CodecMhm1:
		return "MHM1"
	case CodecWebVtt:
		return "WebVTT"
	case CodecTtml:
		return "TTML"
	default:
		return "Unknown"
	}
}

// EncryptionConfig represents encryption configuration for a stream
type EncryptionConfig struct {
	ProtectionScheme string
	CryptByteBlock   int
	SkipByteBlock    int
	PerSampleIvSize  int
	ConstantIv       []byte
	KeyId            []byte
}

// StreamInfo holds stream information
type StreamInfo struct {
	StreamType   StreamType
	TrackId      int
	TimeScale    int32
	Duration     int64  // Duration in time scale units
	Codec        Codec
	CodecString  string
	CodecConfig  []byte
	Language     string
	IsEncrypted  bool
	EncryptionConfig *EncryptionConfig
	
	// Extra data that can be set by derived classes
	ExtraData map[string]interface{}
}

// NewStreamInfo creates a new StreamInfo
func NewStreamInfo(streamType StreamType, trackId int, timeScale int32, duration int64,
	codec Codec, codecString string, codecConfig []byte, language string, isEncrypted bool) *StreamInfo {
	return &StreamInfo{
		StreamType:  streamType,
		TrackId:     trackId,
		TimeScale:   timeScale,
		Duration:    duration,
		Codec:       codec,
		CodecString: codecString,
		CodecConfig: make([]byte, len(codecConfig)),
		Language:    language,
		IsEncrypted: isEncrypted,
		ExtraData:   make(map[string]interface{}),
	}
}

// IsValidConfig returns true if this object has appropriate configuration values
func (si *StreamInfo) IsValidConfig() bool {
	return si.StreamType != StreamUnknown && 
		si.TrackId >= 0 && 
		si.TimeScale > 0 && 
		si.Codec != UnknownCodec
}

// ToString returns a human-readable string describing the stream info
func (si *StreamInfo) ToString() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("type: %s", StreamTypeToString(si.StreamType)))
	parts = append(parts, fmt.Sprintf("codec: %s", CodecToString(si.Codec)))
	
	if si.CodecString != "" {
		parts = append(parts, fmt.Sprintf("codec_string: %s", si.CodecString))
	}
	
	parts = append(parts, fmt.Sprintf("track_id: %d", si.TrackId))
	parts = append(parts, fmt.Sprintf("time_scale: %d", si.TimeScale))
	
	if si.Duration > 0 {
		durationMs := si.Duration * 1000 / int64(si.TimeScale)
		parts = append(parts, fmt.Sprintf("duration: %dms", durationMs))
	}
	
	if si.Language != "" {
		parts = append(parts, fmt.Sprintf("language: %s", si.Language))
	}
	
	if si.IsEncrypted {
		parts = append(parts, "encrypted: true")
		if si.EncryptionConfig != nil {
			parts = append(parts, fmt.Sprintf("protection_scheme: %s", si.EncryptionConfig.ProtectionScheme))
		}
	}
	
	return strings.Join(parts, ", ")
}

// Clone creates a new copy of this stream info
func (si *StreamInfo) Clone() *StreamInfo {
	clone := &StreamInfo{
		StreamType:  si.StreamType,
		TrackId:     si.TrackId,
		TimeScale:   si.TimeScale,
		Duration:    si.Duration,
		Codec:       si.Codec,
		CodecString: si.CodecString,
		CodecConfig: make([]byte, len(si.CodecConfig)),
		Language:    si.Language,
		IsEncrypted: si.IsEncrypted,
		ExtraData:   make(map[string]interface{}),
	}
	
	copy(clone.CodecConfig, si.CodecConfig)
	
	if si.EncryptionConfig != nil {
		clone.EncryptionConfig = &EncryptionConfig{
			ProtectionScheme: si.EncryptionConfig.ProtectionScheme,
			CryptByteBlock:   si.EncryptionConfig.CryptByteBlock,
			SkipByteBlock:    si.EncryptionConfig.SkipByteBlock,
			PerSampleIvSize:  si.EncryptionConfig.PerSampleIvSize,
			ConstantIv:       make([]byte, len(si.EncryptionConfig.ConstantIv)),
			KeyId:            make([]byte, len(si.EncryptionConfig.KeyId)),
		}
		copy(clone.EncryptionConfig.ConstantIv, si.EncryptionConfig.ConstantIv)
		copy(clone.EncryptionConfig.KeyId, si.EncryptionConfig.KeyId)
	}
	
	// Copy extra data
	for k, v := range si.ExtraData {
		clone.ExtraData[k] = v
	}
	
	return clone
}

// GetDurationInSeconds returns the duration in seconds
func (si *StreamInfo) GetDurationInSeconds() float64 {
	if si.TimeScale == 0 {
		return 0
	}
	return float64(si.Duration) / float64(si.TimeScale)
}

// GetTimeBase returns the time base as a time.Duration
func (si *StreamInfo) GetTimeBase() time.Duration {
	if si.TimeScale == 0 {
		return 0
	}
	return time.Second / time.Duration(si.TimeScale)
}

// SetEncryptionConfig sets the encryption configuration
func (si *StreamInfo) SetEncryptionConfig(config *EncryptionConfig) {
	si.EncryptionConfig = config
	si.IsEncrypted = config != nil
}

// IsAudio returns true if this is an audio stream
func (si *StreamInfo) IsAudio() bool {
	return si.StreamType == StreamAudio
}

// IsVideo returns true if this is a video stream
func (si *StreamInfo) IsVideo() bool {
	return si.StreamType == StreamVideo
}

// IsText returns true if this is a text stream
func (si *StreamInfo) IsText() bool {
	return si.StreamType == StreamText
}