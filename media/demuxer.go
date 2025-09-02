// Package media provides demuxing functionality for media containers
package media

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Demuxer extracts streams from media containers
type Demuxer interface {
	Initialize(input string) error
	GetStreams() ([]StreamInfo, error)
	ReadPacket() (*MediaPacket, error)
	Seek(timestamp int64) error
	Close() error
}

// MediaPacket represents a media packet/frame
type MediaPacket struct {
	StreamID    uint32
	Data        []byte
	Timestamp   int64
	Duration    int64
	IsKeyFrame  bool
	StreamType  StreamType
}

// StreamType represents the type of media stream
type StreamType int

const (
	StreamTypeUnknown StreamType = iota
	StreamTypeVideo
	StreamTypeAudio  
	StreamTypeText
)

func (s StreamType) String() string {
	switch s {
	case StreamTypeVideo:
		return "Video"
	case StreamTypeAudio:
		return "Audio"
	case StreamTypeText:
		return "Text"
	default:
		return "Unknown"
	}
}

// MP4Demuxer implements Demuxer for MP4 containers
type MP4Demuxer struct {
	file       *os.File
	container  *MP4Container
	mediaInfo  *MediaInfo
	position   int64
}

// NewMP4Demuxer creates a new MP4 demuxer
func NewMP4Demuxer() *MP4Demuxer {
	return &MP4Demuxer{
		container: NewMP4Container(),
	}
}

// Initialize opens and analyzes the input file
func (d *MP4Demuxer) Initialize(input string) error {
	file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	
	d.file = file
	
	// Parse container to get media information
	info, err := d.container.Parse(file)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to parse container: %w", err)
	}
	
	d.mediaInfo = info
	return nil
}

// GetStreams returns available streams in the media
func (d *MP4Demuxer) GetStreams() ([]StreamInfo, error) {
	if d.mediaInfo == nil {
		return nil, fmt.Errorf("demuxer not initialized")
	}
	
	var streams []StreamInfo
	
	// Add video streams
	for _, vs := range d.mediaInfo.VideoStreams {
		streams = append(streams, vs.StreamInfo)
	}
	
	// Add audio streams  
	for _, as := range d.mediaInfo.AudioStreams {
		streams = append(streams, as.StreamInfo)
	}
	
	// Add text streams
	for _, ts := range d.mediaInfo.TextStreams {
		streams = append(streams, ts.StreamInfo)
	}
	
	return streams, nil
}

// ReadPacket reads the next media packet
func (d *MP4Demuxer) ReadPacket() (*MediaPacket, error) {
	if d.file == nil {
		return nil, fmt.Errorf("demuxer not initialized")
	}
	
	// Simplified packet reading - in a real implementation,
	// this would parse MP4 boxes and extract media samples
	buffer := make([]byte, 4096)
	n, err := d.file.Read(buffer)
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}
	
	packet := &MediaPacket{
		StreamID:   1,
		Data:       buffer[:n],
		Timestamp:  d.position,
		Duration:   40000, // ~40ms for video
		IsKeyFrame: d.position%1000000 == 0, // Key frame every second
		StreamType: StreamTypeVideo,
	}
	
	d.position += 40000
	return packet, nil
}

// Seek seeks to a specific timestamp
func (d *MP4Demuxer) Seek(timestamp int64) error {
	if d.file == nil {
		return fmt.Errorf("demuxer not initialized")
	}
	
	// Simplified seeking - in a real implementation,
	// this would use MP4 index tables
	d.position = timestamp
	return nil
}

// Close closes the demuxer and releases resources
func (d *MP4Demuxer) Close() error {
	if d.file != nil {
		err := d.file.Close()
		d.file = nil
		d.mediaInfo = nil
		return err
	}
	return nil
}

// NewDemuxer creates a demuxer based on file format
func NewDemuxer(input string) (Demuxer, error) {
	ext := filepath.Ext(input)
	
	switch ext {
	case ".mp4", ".m4v", ".m4a":
		return NewMP4Demuxer(), nil
	default:
		// Default to MP4 demuxer
		return NewMP4Demuxer(), nil
	}
}