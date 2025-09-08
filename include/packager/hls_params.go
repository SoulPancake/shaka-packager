// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// HlsPlaylistType defines the EXT-X-PLAYLIST-TYPE in the HLS specification.
// For HlsPlaylistType of Live, EXT-X-PLAYLIST-TYPE tag is omitted.
type HlsPlaylistType int

const (
	HlsPlaylistTypeVod HlsPlaylistType = iota
	HlsPlaylistTypeEvent
	HlsPlaylistTypeLive
)

func (t HlsPlaylistType) String() string {
	switch t {
	case HlsPlaylistTypeVod:
		return "VOD"
	case HlsPlaylistTypeEvent:
		return "EVENT"
	case HlsPlaylistTypeLive:
		return "LIVE"
	default:
		return "UNKNOWN"
	}
}

// HlsParams contains HLS related parameters.
type HlsParams struct {
	// HLS playlist type. See HLS specification for details.
	PlaylistType HlsPlaylistType
	// HLS master playlist output path.
	MasterPlaylistOutput string
	// The base URL for the Media Playlists and media files listed in the
	// playlists. This is the prefix for the files.
	BaseURL string
	// Defines the live window, or the guaranteed duration of the time shifting
	// buffer for 'live' playlists.
	TimeShiftBufferDepth float64
	// Segments outside the live window (defined by 'time_shift_buffer_depth'
	// above) are automatically removed except for the most recent X segments
	// defined by this parameter. This is needed to accommodate latencies in
	// various stages of content serving pipeline, so that the segments stay
	// accessible as they may still be accessed by the player. The segments are
	// not removed if the value is zero.
	PreservedSegmentsOutsideLiveWindow uint64
	// Defines the key uri for "identity" and "com.apple.streamingkeydelivery"
	// key formats. Ignored if the playlist is not encrypted or not using the
	// above key formats.
	KeyURI string
	// The renditions tagged with this language will have 'DEFAULT' set to 'YES'
	// in 'EXT-X-MEDIA' tag. This allows the player to choose the correct default
	// language for the content.
	// This applies to both audio and text tracks. The default language for text
	// tracks can be overridden by 'DefaultTextLanguage'.
	DefaultLanguage string
	// Same as above, but this overrides the default language for text tracks,
	// i.e. subtitles or close-captions.
	DefaultTextLanguage string
	// Indicates that all media samples in the media segments can be decoded
	// without information from other segments.
	IsIndependentSegments bool
	// This is the target segment duration requested by the user. The actual
	// segment duration may be different to the target segment duration. It will
	// be populated from segment duration specified in ChunkingParams if not
	// specified.
	TargetSegmentDuration float64
	// Custom EXT-X-MEDIA-SEQUENCE value to allow continuous media playback
	// across packager restarts. See #691 for details.
	MediaSequenceNumber uint32
	// Sets EXT-X-START on the media playlists to specify the preferred point
	// at which the player should start playing.
	// A positive number indicates a time offset from the beginning of the
	// playlist. A negative number indicates a negative time offset from the end
	// of the last media segment in the playlist.
	StartTimeOffset *float64
	// Create EXT-X-SESSION-KEY in master playlist
	CreateSessionKeys bool
}

// NewHlsParams creates a HlsParams with default values.
func NewHlsParams() HlsParams {
	return HlsParams{
		PlaylistType:                      HlsPlaylistTypeVod,
		TimeShiftBufferDepth:              0,
		PreservedSegmentsOutsideLiveWindow: 0,
		IsIndependentSegments:             false,
		TargetSegmentDuration:             0,
		MediaSequenceNumber:               0,
		CreateSessionKeys:                 false,
	}
}