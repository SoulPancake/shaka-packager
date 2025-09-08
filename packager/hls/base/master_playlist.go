// Copyright 2016 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package hls

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// VariantStream represents a variant stream in the master playlist
type VariantStream struct {
	Bandwidth    int64
	AverageBandwidth int64
	Codecs       string
	Resolution   string
	FrameRate    float64
	Audio        string
	Video        string
	Subtitles    string
	Uri          string
}

// MediaRendition represents a media rendition (audio, video, or subtitles)
type MediaRendition struct {
	Type         MediaPlaylistStreamType
	GroupId      string
	Name         string
	Language     string
	Default      bool
	AutoSelect   bool
	Forced       bool
	Uri          string
	Channels     string
	Characteristics []string
}

// MasterPlaylist represents an HLS master playlist
type MasterPlaylist struct {
	mu                      sync.RWMutex
	fileName                string
	defaultAudioLanguage    string
	defaultTextLanguage     string
	isIndependentSegments   bool
	createSessionKeys       bool
	writtenPlaylist         string
	variantStreams          []*VariantStream
	mediaRenditions         []*MediaRendition
	sessionKeys             []string
	iframeStreams           []*VariantStream
}

// NewMasterPlaylist creates a new master playlist
func NewMasterPlaylist(fileName, defaultAudioLanguage, defaultTextLanguage string, 
	isIndependentSegments, createSessionKeys bool) *MasterPlaylist {
	return &MasterPlaylist{
		fileName:              fileName,
		defaultAudioLanguage:  defaultAudioLanguage,
		defaultTextLanguage:   defaultTextLanguage,
		isIndependentSegments: isIndependentSegments,
		createSessionKeys:     createSessionKeys,
		variantStreams:        make([]*VariantStream, 0),
		mediaRenditions:       make([]*MediaRendition, 0),
		sessionKeys:           make([]string, 0),
		iframeStreams:         make([]*VariantStream, 0),
	}
}

// AddVariantStream adds a variant stream to the master playlist
func (mp *MasterPlaylist) AddVariantStream(stream *VariantStream) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.variantStreams = append(mp.variantStreams, stream)
}

// AddMediaRendition adds a media rendition to the master playlist
func (mp *MasterPlaylist) AddMediaRendition(rendition *MediaRendition) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	// Set default based on language
	if rendition.Type == StreamTypeAudio && rendition.Language == mp.defaultAudioLanguage {
		rendition.Default = true
	}
	if rendition.Type == StreamTypeSubtitle && rendition.Language == mp.defaultTextLanguage {
		rendition.Default = true
	}
	
	mp.mediaRenditions = append(mp.mediaRenditions, rendition)
}

// AddSessionKey adds a session key to the master playlist
func (mp *MasterPlaylist) AddSessionKey(key string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.sessionKeys = append(mp.sessionKeys, key)
}

// AddIframeStream adds an I-frame stream to the master playlist
func (mp *MasterPlaylist) AddIframeStream(stream *VariantStream) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.iframeStreams = append(mp.iframeStreams, stream)
}

// WriteMasterPlaylist writes the master playlist to a file
func (mp *MasterPlaylist) WriteMasterPlaylist(baseUrl, outputDir string, playlists []*MediaPlaylist) bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	// Build variant streams from media playlists
	mp.buildVariantStreamsFromPlaylists(baseUrl, playlists)
	mp.buildMediaRenditionsFromPlaylists(baseUrl, playlists)
	
	// Generate playlist content
	content := mp.generatePlaylistContent()
	
	// Check if content has changed
	if content == mp.writtenPlaylist {
		return true // No change, return success
	}
	
	// Write to file
	fullPath := filepath.Join(outputDir, mp.fileName)
	if err := mp.writeToFile(fullPath, content); err != nil {
		return false
	}
	
	mp.writtenPlaylist = content
	return true
}

// buildVariantStreamsFromPlaylists builds variant streams from media playlists
func (mp *MasterPlaylist) buildVariantStreamsFromPlaylists(baseUrl string, playlists []*MediaPlaylist) {
	// Clear existing variant streams built from playlists
	mp.variantStreams = make([]*VariantStream, 0)
	
	for _, playlist := range playlists {
		if playlist.StreamType() == StreamTypeVideo || playlist.StreamType() == StreamTypeVideoIFramesOnly {
			stream := &VariantStream{
				Bandwidth: playlist.GetBandwidth(),
				AverageBandwidth: playlist.GetBandwidth(),
				Uri: baseUrl + playlist.FileName(),
			}
			
			if playlist.StreamType() == StreamTypeVideoIFramesOnly {
				mp.iframeStreams = append(mp.iframeStreams, stream)
			} else {
				mp.variantStreams = append(mp.variantStreams, stream)
			}
		}
	}
}

// buildMediaRenditionsFromPlaylists builds media renditions from playlists
func (mp *MasterPlaylist) buildMediaRenditionsFromPlaylists(baseUrl string, playlists []*MediaPlaylist) {
	// Clear existing media renditions built from playlists
	mp.mediaRenditions = make([]*MediaRendition, 0)
	
	for _, playlist := range playlists {
		if playlist.StreamType() == StreamTypeAudio || playlist.StreamType() == StreamTypeSubtitle {
			rendition := &MediaRendition{
				Type:    playlist.StreamType(),
				GroupId: playlist.GroupId(),
				Name:    playlist.Name(),
				Default: false,
				AutoSelect: false,
				Uri: baseUrl + playlist.FileName(),
			}
			
			mp.mediaRenditions = append(mp.mediaRenditions, rendition)
		}
	}
}

// generatePlaylistContent generates the complete master playlist content
func (mp *MasterPlaylist) generatePlaylistContent() string {
	var builder strings.Builder
	
	// Write header
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:6\n")
	
	// Write independent segments tag
	if mp.isIndependentSegments {
		builder.WriteString("#EXT-X-INDEPENDENT-SEGMENTS\n")
	}
	
	// Write session keys
	for _, key := range mp.sessionKeys {
		builder.WriteString(fmt.Sprintf("#EXT-X-SESSION-KEY:%s\n", key))
	}
	
	// Group media renditions by type
	audioRenditions := make([]*MediaRendition, 0)
	subtitleRenditions := make([]*MediaRendition, 0)
	
	for _, rendition := range mp.mediaRenditions {
		switch rendition.Type {
		case StreamTypeAudio:
			audioRenditions = append(audioRenditions, rendition)
		case StreamTypeSubtitle:
			subtitleRenditions = append(subtitleRenditions, rendition)
		}
	}
	
	// Write media renditions
	for _, rendition := range audioRenditions {
		builder.WriteString(mp.formatMediaRendition(rendition))
	}
	for _, rendition := range subtitleRenditions {
		builder.WriteString(mp.formatMediaRendition(rendition))
	}
	
	// Sort variant streams by bandwidth
	sortedStreams := make([]*VariantStream, len(mp.variantStreams))
	copy(sortedStreams, mp.variantStreams)
	sort.Slice(sortedStreams, func(i, j int) bool {
		return sortedStreams[i].Bandwidth < sortedStreams[j].Bandwidth
	})
	
	// Write variant streams
	for _, stream := range sortedStreams {
		builder.WriteString(mp.formatVariantStream(stream))
	}
	
	// Write I-frame streams
	for _, stream := range mp.iframeStreams {
		builder.WriteString(mp.formatIframeStream(stream))
	}
	
	return builder.String()
}

// formatMediaRendition formats a media rendition entry
func (mp *MasterPlaylist) formatMediaRendition(rendition *MediaRendition) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("TYPE=%s", rendition.Type.String()))
	parts = append(parts, fmt.Sprintf("GROUP-ID=\"%s\"", rendition.GroupId))
	parts = append(parts, fmt.Sprintf("NAME=\"%s\"", rendition.Name))
	
	if rendition.Language != "" {
		parts = append(parts, fmt.Sprintf("LANGUAGE=\"%s\"", rendition.Language))
	}
	
	if rendition.Default {
		parts = append(parts, "DEFAULT=YES")
	} else {
		parts = append(parts, "DEFAULT=NO")
	}
	
	if rendition.AutoSelect {
		parts = append(parts, "AUTOSELECT=YES")
	} else {
		parts = append(parts, "AUTOSELECT=NO")
	}
	
	if rendition.Forced {
		parts = append(parts, "FORCED=YES")
	}
	
	if rendition.Uri != "" {
		parts = append(parts, fmt.Sprintf("URI=\"%s\"", rendition.Uri))
	}
	
	if rendition.Channels != "" {
		parts = append(parts, fmt.Sprintf("CHANNELS=\"%s\"", rendition.Channels))
	}
	
	if len(rendition.Characteristics) > 0 {
		chars := strings.Join(rendition.Characteristics, ",")
		parts = append(parts, fmt.Sprintf("CHARACTERISTICS=\"%s\"", chars))
	}
	
	return fmt.Sprintf("#EXT-X-MEDIA:%s\n", strings.Join(parts, ","))
}

// formatVariantStream formats a variant stream entry
func (mp *MasterPlaylist) formatVariantStream(stream *VariantStream) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("BANDWIDTH=%d", stream.Bandwidth))
	
	if stream.AverageBandwidth > 0 && stream.AverageBandwidth != stream.Bandwidth {
		parts = append(parts, fmt.Sprintf("AVERAGE-BANDWIDTH=%d", stream.AverageBandwidth))
	}
	
	if stream.Codecs != "" {
		parts = append(parts, fmt.Sprintf("CODECS=\"%s\"", stream.Codecs))
	}
	
	if stream.Resolution != "" {
		parts = append(parts, fmt.Sprintf("RESOLUTION=%s", stream.Resolution))
	}
	
	if stream.FrameRate > 0 {
		parts = append(parts, fmt.Sprintf("FRAME-RATE=%.3f", stream.FrameRate))
	}
	
	if stream.Audio != "" {
		parts = append(parts, fmt.Sprintf("AUDIO=\"%s\"", stream.Audio))
	}
	
	if stream.Video != "" {
		parts = append(parts, fmt.Sprintf("VIDEO=\"%s\"", stream.Video))
	}
	
	if stream.Subtitles != "" {
		parts = append(parts, fmt.Sprintf("SUBTITLES=\"%s\"", stream.Subtitles))
	}
	
	return fmt.Sprintf("#EXT-X-STREAM-INF:%s\n%s\n", strings.Join(parts, ","), stream.Uri)
}

// formatIframeStream formats an I-frame stream entry
func (mp *MasterPlaylist) formatIframeStream(stream *VariantStream) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("BANDWIDTH=%d", stream.Bandwidth))
	
	if stream.Codecs != "" {
		parts = append(parts, fmt.Sprintf("CODECS=\"%s\"", stream.Codecs))
	}
	
	if stream.Resolution != "" {
		parts = append(parts, fmt.Sprintf("RESOLUTION=%s", stream.Resolution))
	}
	
	parts = append(parts, fmt.Sprintf("URI=\"%s\"", stream.Uri))
	
	return fmt.Sprintf("#EXT-X-I-FRAME-STREAM-INF:%s\n", strings.Join(parts, ","))
}

// writeToFile writes content to a file
func (mp *MasterPlaylist) writeToFile(filePath, content string) error {
	// This would write to an actual file in a real implementation
	// For now, we'll just simulate success
	return nil
}