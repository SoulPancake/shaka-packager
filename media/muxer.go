// Package media provides muxing functionality for creating media containers
package media

import (
	"fmt"
	"os"
	"path/filepath"
)

// Muxer combines media streams into containers
type Muxer interface {
	Initialize(output string, streams []StreamDescriptor) error
	WritePacket(packet *MediaPacket) error
	Finalize() error
	Close() error
}

// StreamDescriptor describes an output stream
type StreamDescriptor struct {
	StreamID   uint32
	StreamType StreamType
	Codec      string
	Bitrate    int64
	Language   string
	ExtraData  []byte
}

// MP4Muxer implements Muxer for MP4 output
type MP4Muxer struct {
	file        *os.File
	container   *MP4Container
	streams     []StreamDescriptor
	packets     []*MediaPacket
	initialized bool
	closed      bool
}

// NewMP4Muxer creates a new MP4 muxer
func NewMP4Muxer() *MP4Muxer {
	return &MP4Muxer{
		container: NewMP4Container(),
		packets:   make([]*MediaPacket, 0),
	}
}

// Initialize sets up the muxer for output
func (m *MP4Muxer) Initialize(output string, streams []StreamDescriptor) error {
	if m.initialized {
		return fmt.Errorf("muxer already initialized")
	}

	// Create output directory if needed
	dir := filepath.Dir(output)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	m.file = file
	m.streams = make([]StreamDescriptor, len(streams))
	copy(m.streams, streams)
	m.initialized = true

	return nil
}

// WritePacket writes a media packet to the output
func (m *MP4Muxer) WritePacket(packet *MediaPacket) error {
	if !m.initialized {
		return fmt.Errorf("muxer not initialized")
	}
	if m.closed {
		return fmt.Errorf("muxer is closed")
	}

	// Store packet for later processing
	packetCopy := &MediaPacket{
		StreamID:   packet.StreamID,
		Data:       make([]byte, len(packet.Data)),
		Timestamp:  packet.Timestamp,
		Duration:   packet.Duration,
		IsKeyFrame: packet.IsKeyFrame,
		StreamType: packet.StreamType,
	}
	copy(packetCopy.Data, packet.Data)
	
	m.packets = append(m.packets, packetCopy)
	return nil
}

// Finalize completes the output file
func (m *MP4Muxer) Finalize() error {
	if !m.initialized {
		return fmt.Errorf("muxer not initialized")
	}
	if m.closed {
		return fmt.Errorf("muxer is closed")
	}

	// Create MediaInfo from collected data
	info := &MediaInfo{
		Format:    FormatMP4,
		Timescale: 90000,
	}

	// Add streams based on collected packets
	for _, stream := range m.streams {
		switch stream.StreamType {
		case StreamTypeVideo:
			info.VideoStreams = append(info.VideoStreams, VideoStream{
				StreamInfo: StreamInfo{
					ID:      stream.StreamID,
					Bitrate: stream.Bitrate,
				},
				Codec:     stream.Codec,
				Width:     1920,
				Height:    1080,
				FrameRate: 30.0,
			})
		case StreamTypeAudio:
			info.AudioStreams = append(info.AudioStreams, AudioStream{
				StreamInfo: StreamInfo{
					ID:      stream.StreamID,
					Bitrate: stream.Bitrate,
				},
				Codec:      stream.Codec,
				Channels:   2,
				SampleRate: 44100,
			})
		case StreamTypeText:
			info.TextStreams = append(info.TextStreams, TextStream{
				StreamInfo: StreamInfo{
					ID:      stream.StreamID,
					Bitrate: stream.Bitrate,
				},
				Codec: stream.Codec,
			})
		}
	}

	// Write the container structure
	if err := m.container.Write(m.file, info); err != nil {
		return fmt.Errorf("failed to write container: %w", err)
	}

	// Write media data
	for _, packet := range m.packets {
		if _, err := m.file.Write(packet.Data); err != nil {
			return fmt.Errorf("failed to write packet data: %w", err)
		}
	}

	return nil
}

// Close closes the muxer and releases resources
func (m *MP4Muxer) Close() error {
	if m.closed {
		return nil
	}

	var err error
	if m.file != nil {
		err = m.file.Close()
		m.file = nil
	}
	
	m.closed = true
	m.initialized = false
	m.packets = nil
	m.streams = nil
	
	return err
}

// SegmentedMuxer handles segmented output (for HLS/DASH)
type SegmentedMuxer struct {
	baseMuxer     Muxer
	segmentLength int64
	segmentIndex  int
	currentTime   int64
	outputPattern string
	currentFile   string
}

// NewSegmentedMuxer creates a new segmented muxer
func NewSegmentedMuxer(outputPattern string, segmentLength int64) *SegmentedMuxer {
	return &SegmentedMuxer{
		outputPattern: outputPattern,
		segmentLength: segmentLength,
		segmentIndex:  0,
		currentTime:   0,
	}
}

// Initialize sets up the segmented muxer
func (sm *SegmentedMuxer) Initialize(output string, streams []StreamDescriptor) error {
	sm.outputPattern = output
	return sm.createNewSegment(streams)
}

// WritePacket writes a packet, creating new segments as needed
func (sm *SegmentedMuxer) WritePacket(packet *MediaPacket) error {
	// Check if we need a new segment
	if packet.Timestamp >= sm.currentTime+sm.segmentLength {
		if err := sm.baseMuxer.Finalize(); err != nil {
			return err
		}
		if err := sm.baseMuxer.Close(); err != nil {
			return err
		}
		
		sm.segmentIndex++
		sm.currentTime = packet.Timestamp
		
		// Create new segment - need streams from previous segment
		// This is simplified - real implementation would track stream info
		streams := []StreamDescriptor{
			{StreamID: 1, StreamType: StreamTypeVideo, Codec: "avc1"},
			{StreamID: 2, StreamType: StreamTypeAudio, Codec: "mp4a"},
		}
		if err := sm.createNewSegment(streams); err != nil {
			return err
		}
	}
	
	return sm.baseMuxer.WritePacket(packet)
}

// Finalize completes the current segment
func (sm *SegmentedMuxer) Finalize() error {
	if sm.baseMuxer != nil {
		return sm.baseMuxer.Finalize()
	}
	return nil
}

// Close closes all resources
func (sm *SegmentedMuxer) Close() error {
	if sm.baseMuxer != nil {
		return sm.baseMuxer.Close()
	}
	return nil
}

// createNewSegment creates a new output segment
func (sm *SegmentedMuxer) createNewSegment(streams []StreamDescriptor) error {
	filename := fmt.Sprintf(sm.outputPattern, sm.segmentIndex)
	sm.currentFile = filename
	
	sm.baseMuxer = NewMP4Muxer()
	return sm.baseMuxer.Initialize(filename, streams)
}

// NewMuxer creates a muxer based on output format
func NewMuxer(output string, segmented bool, segmentLength int64) Muxer {
	if segmented {
		return NewSegmentedMuxer(output, segmentLength)
	}
	
	ext := filepath.Ext(output)
	switch ext {
	case ".mp4", ".m4v", ".m4a":
		return NewMP4Muxer()
	default:
		return NewMP4Muxer() // Default to MP4
	}
}