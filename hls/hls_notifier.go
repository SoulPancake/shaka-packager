// Package hls provides HLS (HTTP Live Streaming) functionality
package hls

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HLSNotifier handles HLS playlist generation and updates
type HLSNotifier interface {
	NotifyNewStream(streamInfo StreamInfo) error
	NotifyNewSegment(streamID uint32, segmentInfo SegmentInfo) error
	NotifyEncryptionUpdate(streamID uint32, keyInfo KeyInfo) error
	Flush() error
	Close() error
}

// StreamInfo contains information about an HLS stream
type StreamInfo struct {
	StreamID    uint32
	StreamType  StreamType
	Codec       string
	Bandwidth   uint64
	Resolution  string
	FrameRate   float64
	AudioCodec  string
	Channels    uint32
	SampleRate  uint32
	Language    string
	PlaylistName string
}

// StreamType represents the type of HLS stream
type StreamType int

const (
	StreamTypeUnknown StreamType = iota
	StreamTypeVideo
	StreamTypeAudio
	StreamTypeSubtitle
)

// SegmentInfo contains information about an HLS segment
type SegmentInfo struct {
	SegmentID    uint32
	Filename     string
	Duration     time.Duration
	Size         uint64
	Timestamp    int64
	IsDiscontinuity bool
	IsKeyFrame   bool
}

// KeyInfo contains encryption key information
type KeyInfo struct {
	Method    string
	URI       string
	IV        []byte
	KeyFormat string
}

// SimpleHLSNotifier implements basic HLS playlist generation
type SimpleHLSNotifier struct {
	outputDir     string
	masterPlaylist *MasterPlaylist
	streamPlaylists map[uint32]*MediaPlaylist
	segmentList    map[uint32][]SegmentInfo
	keyInfo       map[uint32]KeyInfo
}

// NewSimpleHLSNotifier creates a new simple HLS notifier
func NewSimpleHLSNotifier(outputDir string) *SimpleHLSNotifier {
	return &SimpleHLSNotifier{
		outputDir:       outputDir,
		masterPlaylist:  NewMasterPlaylist(),
		streamPlaylists: make(map[uint32]*MediaPlaylist),
		segmentList:     make(map[uint32][]SegmentInfo),
		keyInfo:        make(map[uint32]KeyInfo),
	}
}

// NotifyNewStream adds a new stream to the HLS output
func (h *SimpleHLSNotifier) NotifyNewStream(streamInfo StreamInfo) error {
	// Add to master playlist
	variant := PlaylistVariant{
		Bandwidth:   streamInfo.Bandwidth,
		Codec:       streamInfo.Codec,
		Resolution:  streamInfo.Resolution,
		FrameRate:   streamInfo.FrameRate,
		AudioCodec:  streamInfo.AudioCodec,
		URI:         streamInfo.PlaylistName,
	}
	
	h.masterPlaylist.AddVariant(variant)
	
	// Create media playlist for this stream
	mediaPlaylist := NewMediaPlaylist(streamInfo.StreamID)
	h.streamPlaylists[streamInfo.StreamID] = mediaPlaylist
	h.segmentList[streamInfo.StreamID] = make([]SegmentInfo, 0)
	
	return nil
}

// NotifyNewSegment adds a new segment to the HLS playlist
func (h *SimpleHLSNotifier) NotifyNewSegment(streamID uint32, segmentInfo SegmentInfo) error {
	playlist, exists := h.streamPlaylists[streamID]
	if !exists {
		return fmt.Errorf("stream %d not found", streamID)
	}
	
	// Add segment to playlist
	segment := Segment{
		Filename:        segmentInfo.Filename,
		Duration:        segmentInfo.Duration,
		Size:           segmentInfo.Size,
		IsDiscontinuity: segmentInfo.IsDiscontinuity,
	}
	
	// Add encryption info if available
	if keyInfo, hasKey := h.keyInfo[streamID]; hasKey {
		segment.KeyInfo = &keyInfo
	}
	
	playlist.AddSegment(segment)
	
	// Store segment info
	h.segmentList[streamID] = append(h.segmentList[streamID], segmentInfo)
	
	return nil
}

// NotifyEncryptionUpdate updates encryption information
func (h *SimpleHLSNotifier) NotifyEncryptionUpdate(streamID uint32, keyInfo KeyInfo) error {
	h.keyInfo[streamID] = keyInfo
	return nil
}

// Flush writes all playlists to disk
func (h *SimpleHLSNotifier) Flush() error {
	// Write master playlist
	masterPath := filepath.Join(h.outputDir, "playlist.m3u8")
	if err := h.writeMasterPlaylist(masterPath); err != nil {
		return fmt.Errorf("failed to write master playlist: %w", err)
	}
	
	// Write media playlists
	for streamID, playlist := range h.streamPlaylists {
		filename := fmt.Sprintf("stream_%d.m3u8", streamID)
		playlistPath := filepath.Join(h.outputDir, filename)
		if err := h.writeMediaPlaylist(playlistPath, playlist); err != nil {
			return fmt.Errorf("failed to write media playlist %s: %w", filename, err)
		}
	}
	
	return nil
}

// Close cleans up resources
func (h *SimpleHLSNotifier) Close() error {
	return h.Flush()
}

// writeMasterPlaylist writes the master playlist to file
func (h *SimpleHLSNotifier) writeMasterPlaylist(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return h.masterPlaylist.WriteTo(file)
}

// writeMediaPlaylist writes a media playlist to file
func (h *SimpleHLSNotifier) writeMediaPlaylist(filename string, playlist *MediaPlaylist) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return playlist.WriteTo(file)
}

// MasterPlaylist represents an HLS master playlist
type MasterPlaylist struct {
	Variants []PlaylistVariant
}

// PlaylistVariant represents a variant in the master playlist
type PlaylistVariant struct {
	Bandwidth   uint64
	Codec       string
	Resolution  string
	FrameRate   float64
	AudioCodec  string
	URI         string
}

// NewMasterPlaylist creates a new master playlist
func NewMasterPlaylist() *MasterPlaylist {
	return &MasterPlaylist{
		Variants: make([]PlaylistVariant, 0),
	}
}

// AddVariant adds a variant to the master playlist
func (m *MasterPlaylist) AddVariant(variant PlaylistVariant) {
	m.Variants = append(m.Variants, variant)
	
	// Sort by bandwidth
	sort.Slice(m.Variants, func(i, j int) bool {
		return m.Variants[i].Bandwidth < m.Variants[j].Bandwidth
	})
}

// WriteTo writes the master playlist to a writer
func (m *MasterPlaylist) WriteTo(w io.Writer) error {
	// Write header
	if _, err := w.Write([]byte("#EXTM3U\n")); err != nil {
		return err
	}
	if _, err := w.Write([]byte("#EXT-X-VERSION:3\n")); err != nil {
		return err
	}
	
	// Write variants
	for _, variant := range m.Variants {
		line := fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d", variant.Bandwidth)
		
		if variant.Resolution != "" {
			line += fmt.Sprintf(",RESOLUTION=%s", variant.Resolution)
		}
		if variant.FrameRate > 0 {
			line += fmt.Sprintf(",FRAME-RATE=%.3f", variant.FrameRate)
		}
		if variant.Codec != "" {
			line += fmt.Sprintf(",CODECS=\"%s\"", variant.Codec)
		}
		if variant.AudioCodec != "" && variant.AudioCodec != variant.Codec {
			if variant.Codec != "" {
				line = strings.Replace(line, fmt.Sprintf("CODECS=\"%s\"", variant.Codec), 
					fmt.Sprintf("CODECS=\"%s,%s\"", variant.Codec, variant.AudioCodec), 1)
			} else {
				line += fmt.Sprintf(",CODECS=\"%s\"", variant.AudioCodec)
			}
		}
		
		line += "\n"
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}
		
		if _, err := w.Write([]byte(variant.URI + "\n")); err != nil {
			return err
		}
	}
	
	return nil
}

// MediaPlaylist represents an HLS media playlist
type MediaPlaylist struct {
	StreamID        uint32
	Segments        []Segment
	TargetDuration  time.Duration
	MediaSequence   uint64
	IsLive         bool
}

// Segment represents an HLS media segment
type Segment struct {
	Filename        string
	Duration        time.Duration
	Size           uint64
	IsDiscontinuity bool
	KeyInfo        *KeyInfo
}

// NewMediaPlaylist creates a new media playlist
func NewMediaPlaylist(streamID uint32) *MediaPlaylist {
	return &MediaPlaylist{
		StreamID:       streamID,
		Segments:       make([]Segment, 0),
		TargetDuration: 6 * time.Second,
		MediaSequence:  0,
		IsLive:        false,
	}
}

// AddSegment adds a segment to the media playlist
func (m *MediaPlaylist) AddSegment(segment Segment) {
	m.Segments = append(m.Segments, segment)
	
	// Update target duration
	if segment.Duration > m.TargetDuration {
		m.TargetDuration = segment.Duration
	}
}

// WriteTo writes the media playlist to a writer
func (m *MediaPlaylist) WriteTo(w io.Writer) error {
	// Write header
	if _, err := w.Write([]byte("#EXTM3U\n")); err != nil {
		return err
	}
	if _, err := w.Write([]byte("#EXT-X-VERSION:3\n")); err != nil {
		return err
	}
	
	// Write target duration
	line := fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", int(m.TargetDuration.Seconds()))
	if _, err := w.Write([]byte(line)); err != nil {
		return err
	}
	
	// Write media sequence
	line = fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", m.MediaSequence)
	if _, err := w.Write([]byte(line)); err != nil {
		return err
	}
	
	// Write segments
	for _, segment := range m.Segments {
		// Write discontinuity tag if needed
		if segment.IsDiscontinuity {
			if _, err := w.Write([]byte("#EXT-X-DISCONTINUITY\n")); err != nil {
				return err
			}
		}
		
		// Write key info if present
		if segment.KeyInfo != nil {
			line := fmt.Sprintf("#EXT-X-KEY:METHOD=%s,URI=\"%s\"", 
				segment.KeyInfo.Method, segment.KeyInfo.URI)
			if len(segment.KeyInfo.IV) > 0 {
				line += fmt.Sprintf(",IV=0x%X", segment.KeyInfo.IV)
			}
			line += "\n"
			if _, err := w.Write([]byte(line)); err != nil {
				return err
			}
		}
		
		// Write segment info
		line := fmt.Sprintf("#EXTINF:%.6f,\n", segment.Duration.Seconds())
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}
		
		if _, err := w.Write([]byte(segment.Filename + "\n")); err != nil {
			return err
		}
	}
	
	// Write end tag for VOD
	if !m.IsLive {
		if _, err := w.Write([]byte("#EXT-X-ENDLIST\n")); err != nil {
			return err
		}
	}
	
	return nil
}