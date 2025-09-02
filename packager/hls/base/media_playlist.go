// Copyright 2016 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package hls

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

// HlsEntry represents different types of HLS playlist entries
type HlsEntry interface {
	Type() EntryType
	ToString() string
}

// EntryType represents the type of HLS entry
type EntryType int

const (
	ExtInf EntryType = iota
	ExtKey
	ExtDiscontinuity
	ExtPlacementOpportunity
)

// ExtInfEntry represents an #EXTINF entry
type ExtInfEntry struct {
	Duration float64
	Title    string
	Uri      string
}

// Type returns the entry type
func (e *ExtInfEntry) Type() EntryType {
	return ExtInf
}

// ToString returns the string representation
func (e *ExtInfEntry) ToString() string {
	return fmt.Sprintf("#EXTINF:%.6f,%s\n%s", e.Duration, e.Title, e.Uri)
}

// ExtKeyEntry represents an #EXT-X-KEY entry
type ExtKeyEntry struct {
	Method EncryptionMethod
	Uri    string
	IV     string
	KeyFormat string
	KeyFormatVersions string
}

// Type returns the entry type
func (e *ExtKeyEntry) Type() EntryType {
	return ExtKey
}

// ToString returns the string representation
func (e *ExtKeyEntry) ToString() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("METHOD=%s", e.Method.String()))
	
	if e.Uri != "" {
		parts = append(parts, fmt.Sprintf("URI=\"%s\"", e.Uri))
	}
	if e.IV != "" {
		parts = append(parts, fmt.Sprintf("IV=%s", e.IV))
	}
	if e.KeyFormat != "" {
		parts = append(parts, fmt.Sprintf("KEYFORMAT=\"%s\"", e.KeyFormat))
	}
	if e.KeyFormatVersions != "" {
		parts = append(parts, fmt.Sprintf("KEYFORMATVERSIONS=\"%s\"", e.KeyFormatVersions))
	}
	
	return fmt.Sprintf("#EXT-X-KEY:%s", strings.Join(parts, ","))
}

// ExtDiscontinuityEntry represents an #EXT-X-DISCONTINUITY entry
type ExtDiscontinuityEntry struct{}

// Type returns the entry type
func (e *ExtDiscontinuityEntry) Type() EntryType {
	return ExtDiscontinuity
}

// ToString returns the string representation
func (e *ExtDiscontinuityEntry) ToString() string {
	return "#EXT-X-DISCONTINUITY"
}

// ExtPlacementOpportunityEntry represents an #EXT-X-PLACEMENT-OPPORTUNITY entry
type ExtPlacementOpportunityEntry struct{}

// Type returns the entry type
func (e *ExtPlacementOpportunityEntry) Type() EntryType {
	return ExtPlacementOpportunity
}

// ToString returns the string representation
func (e *ExtPlacementOpportunityEntry) ToString() string {
	return "#EXT-X-PLACEMENT-OPPORTUNITY"
}

// MediaPlaylistStreamType represents the type of media stream
type MediaPlaylistStreamType int

const (
	StreamTypeUnknown MediaPlaylistStreamType = iota
	StreamTypeAudio
	StreamTypeVideo
	StreamTypeVideoIFramesOnly
	StreamTypeSubtitle
)

// String returns the string representation of the stream type
func (st MediaPlaylistStreamType) String() string {
	switch st {
	case StreamTypeAudio:
		return "AUDIO"
	case StreamTypeVideo:
		return "VIDEO"
	case StreamTypeVideoIFramesOnly:
		return "VIDEO"
	case StreamTypeSubtitle:
		return "SUBTITLES"
	default:
		return "UNKNOWN"
	}
}

// EncryptionMethod represents HLS encryption methods
type EncryptionMethod int

const (
	EncryptionNone EncryptionMethod = iota
	EncryptionAes128
	EncryptionSampleAes
	EncryptionSampleAesCenc
)

// String returns the string representation of the encryption method
func (em EncryptionMethod) String() string {
	switch em {
	case EncryptionNone:
		return "NONE"
	case EncryptionAes128:
		return "AES-128"
	case EncryptionSampleAes:
		return "SAMPLE-AES"
	case EncryptionSampleAesCenc:
		return "SAMPLE-AES-CENC"
	default:
		return "NONE"
	}
}

// MediaPlaylist represents an HLS media playlist
type MediaPlaylist struct {
	mu                    sync.RWMutex
	hlsParams             packager.HlsParams
	fileName              string
	name                  string
	groupId               string
	streamType            MediaPlaylistStreamType
	targetDuration        time.Duration
	entries               []HlsEntry
	sequenceNumber        int64
	isLive                bool
	allowCache            bool
	playlistType          string
	encryptionMethod      EncryptionMethod
	discontinuitySequence int64
}

// NewMediaPlaylist creates a new media playlist
func NewMediaPlaylist(hlsParams packager.HlsParams, fileName, name, groupId string) *MediaPlaylist {
	return &MediaPlaylist{
		hlsParams:      hlsParams,
		fileName:       fileName,
		name:           name,
		groupId:        groupId,
		streamType:     StreamTypeUnknown,
		targetDuration: time.Duration(hlsParams.TargetDuration) * time.Second,
		entries:        make([]HlsEntry, 0),
		sequenceNumber: 0,
		allowCache:     true,
		playlistType:   "VOD",
	}
}

// FileName returns the file name of this media playlist
func (mp *MediaPlaylist) FileName() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.fileName
}

// Name returns the name of this playlist
func (mp *MediaPlaylist) Name() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.name
}

// GroupId returns the group ID for this playlist
func (mp *MediaPlaylist) GroupId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.groupId
}

// StreamType returns the stream type
func (mp *MediaPlaylist) StreamType() MediaPlaylistStreamType {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.streamType
}

// SetStreamType sets the stream type
func (mp *MediaPlaylist) SetStreamType(streamType MediaPlaylistStreamType) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.streamType = streamType
}

// AddEntry adds an entry to the playlist
func (mp *MediaPlaylist) AddEntry(entry HlsEntry) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.entries = append(mp.entries, entry)
}

// AddSegment adds a segment to the playlist
func (mp *MediaPlaylist) AddSegment(uri string, duration float64, title string) {
	entry := &ExtInfEntry{
		Duration: duration,
		Title:    title,
		Uri:      uri,
	}
	mp.AddEntry(entry)
}

// AddEncryptionInfo adds encryption information to the playlist
func (mp *MediaPlaylist) AddEncryptionInfo(method EncryptionMethod, uri, iv, keyFormat, keyFormatVersions string) {
	entry := &ExtKeyEntry{
		Method:            method,
		Uri:               uri,
		IV:                iv,
		KeyFormat:         keyFormat,
		KeyFormatVersions: keyFormatVersions,
	}
	mp.AddEntry(entry)
}

// AddDiscontinuity adds a discontinuity marker
func (mp *MediaPlaylist) AddDiscontinuity() {
	mp.AddEntry(&ExtDiscontinuityEntry{})
}

// AddPlacementOpportunity adds a placement opportunity marker
func (mp *MediaPlaylist) AddPlacementOpportunity() {
	mp.AddEntry(&ExtPlacementOpportunityEntry{})
}

// SetTargetDuration sets the target duration for segments
func (mp *MediaPlaylist) SetTargetDuration(duration time.Duration) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.targetDuration = duration
}

// SetSequenceNumber sets the media sequence number
func (mp *MediaPlaylist) SetSequenceNumber(seqNum int64) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.sequenceNumber = seqNum
}

// SetPlaylistType sets the playlist type (VOD or EVENT)
func (mp *MediaPlaylist) SetPlaylistType(playlistType string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.playlistType = playlistType
}

// SetIsLive sets whether this is a live playlist
func (mp *MediaPlaylist) SetIsLive(isLive bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.isLive = isLive
	if isLive {
		mp.playlistType = "EVENT"
	} else {
		mp.playlistType = "VOD"
	}
}

// ToString generates the complete playlist string
func (mp *MediaPlaylist) ToString() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	var builder strings.Builder
	
	// Write header
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:6\n")
	
	// Write target duration
	builder.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", int(mp.targetDuration.Seconds())))
	
	// Write sequence number
	if mp.sequenceNumber > 0 {
		builder.WriteString(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", mp.sequenceNumber))
	}
	
	// Write discontinuity sequence
	if mp.discontinuitySequence > 0 {
		builder.WriteString(fmt.Sprintf("#EXT-X-DISCONTINUITY-SEQUENCE:%d\n", mp.discontinuitySequence))
	}
	
	// Write playlist type
	if mp.playlistType != "" && !mp.isLive {
		builder.WriteString(fmt.Sprintf("#EXT-X-PLAYLIST-TYPE:%s\n", mp.playlistType))
	}
	
	// Write entries
	for _, entry := range mp.entries {
		builder.WriteString(entry.ToString())
		builder.WriteString("\n")
	}
	
	// Write end tag for VOD
	if !mp.isLive {
		builder.WriteString("#EXT-X-ENDLIST\n")
	}
	
	return builder.String()
}

// GetBandwidth returns the bandwidth estimate for this playlist
func (mp *MediaPlaylist) GetBandwidth() int64 {
	// This would typically be calculated based on segment sizes and durations
	// For now, return a default value
	return 1000000 // 1 Mbps default
}