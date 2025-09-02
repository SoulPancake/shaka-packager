// Package media provides core media container and stream processing functionality.
package media

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// MediaContainer represents a media container format (MP4, WebM, etc.)
type MediaContainer interface {
	Parse(r io.Reader) (*MediaInfo, error)
	Write(w io.Writer, info *MediaInfo) error
	GetFormat() ContainerFormat
}

// ContainerFormat represents supported container formats
type ContainerFormat int

const (
	FormatUnknown ContainerFormat = iota
	FormatMP4
	FormatWebM
	FormatTS
	FormatFMP4
)

func (f ContainerFormat) String() string {
	switch f {
	case FormatMP4:
		return "MP4"
	case FormatWebM:
		return "WebM"
	case FormatTS:
		return "TS"
	case FormatFMP4:
		return "FMP4"
	default:
		return "Unknown"
	}
}

// MediaInfo contains metadata about a media file
type MediaInfo struct {
	Format        ContainerFormat
	Duration      time.Duration
	Bitrate       int64
	VideoStreams  []VideoStream
	AudioStreams  []AudioStream
	TextStreams   []TextStream
	Timescale     uint32
	CreationTime  time.Time
	ModifiedTime  time.Time
}

// StreamInfo contains common stream information
type StreamInfo struct {
	ID        uint32
	Language  string
	Duration  time.Duration
	Bitrate   int64
	Timescale uint32
}

// VideoStream represents a video stream
type VideoStream struct {
	StreamInfo
	Codec       string
	Width       uint32
	Height      uint32
	FrameRate   float64
	ProfileIDC  uint8
	LevelIDC    uint8
	ChromaFormat uint8
}

// AudioStream represents an audio stream
type AudioStream struct {
	StreamInfo
	Codec          string
	Channels       uint32
	SampleRate     uint32
	BitsPerSample  uint32
	ChannelLayout  uint64
}

// TextStream represents a text/subtitle stream
type TextStream struct {
	StreamInfo
	Codec    string
	MimeType string
}

// MP4Container implements MediaContainer for MP4 format
type MP4Container struct{}

// NewMP4Container creates a new MP4 container parser
func NewMP4Container() *MP4Container {
	return &MP4Container{}
}

// Parse parses an MP4 file and extracts media information
func (mp4 *MP4Container) Parse(r io.Reader) (*MediaInfo, error) {
	info := &MediaInfo{
		Format:    FormatMP4,
		Timescale: 90000, // Default timescale
	}

	// Read MP4 boxes to extract media information
	for {
		var size uint32
		var boxType [4]byte

		// Read box header
		err := binary.Read(r, binary.BigEndian, &size)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read box size: %w", err)
		}

		_, err = io.ReadFull(r, boxType[:])
		if err != nil {
			return nil, fmt.Errorf("failed to read box type: %w", err)
		}

		boxTypeStr := string(boxType[:])
		
		// Handle different box types
		switch boxTypeStr {
		case "moov":
			// Movie header box - contains media metadata
			info.Duration = 60 * time.Second // Placeholder
			info.Bitrate = 1000000 // Placeholder
		case "mvhd":
			// Movie header - parse duration and timescale
			info.CreationTime = time.Now()
			info.ModifiedTime = time.Now()
		case "trak":
			// Track box - contains stream information
			// Add placeholder streams
			info.VideoStreams = append(info.VideoStreams, VideoStream{
				StreamInfo: StreamInfo{
					ID:       1,
					Duration: info.Duration,
					Bitrate:  800000,
				},
				Codec:     "avc1",
				Width:     1920,
				Height:    1080,
				FrameRate: 30.0,
			})
			info.AudioStreams = append(info.AudioStreams, AudioStream{
				StreamInfo: StreamInfo{
					ID:       2,
					Duration: info.Duration,
					Bitrate:  128000,
				},
				Codec:      "mp4a",
				Channels:   2,
				SampleRate: 44100,
			})
		}

		// Skip to next box
		if size > 8 {
			remaining := int64(size - 8)
			if remaining > 0 {
				io.CopyN(io.Discard, r, remaining)
			}
		}
	}

	return info, nil
}

// Write writes media information to MP4 format
func (mp4 *MP4Container) Write(w io.Writer, info *MediaInfo) error {
	// Write basic MP4 structure
	
	// Write ftyp box
	ftypBox := []byte{
		0x00, 0x00, 0x00, 0x18, // size
		'f', 't', 'y', 'p',     // type
		'i', 's', 'o', 'm',     // major brand
		0x00, 0x00, 0x02, 0x00, // minor version
		'i', 's', 'o', 'm',     // compatible brand
		'm', 'p', '4', '1',     // compatible brand
	}
	
	if _, err := w.Write(ftypBox); err != nil {
		return fmt.Errorf("failed to write ftyp box: %w", err)
	}

	// Write moov box header
	moovHeader := []byte{
		0x00, 0x00, 0x00, 0x40, // size (placeholder)
		'm', 'o', 'o', 'v',     // type
	}
	
	if _, err := w.Write(moovHeader); err != nil {
		return fmt.Errorf("failed to write moov box: %w", err)
	}

	// Write mvhd box
	mvhdBox := make([]byte, 56) // Simplified movie header
	copy(mvhdBox[:4], []byte{0x00, 0x00, 0x00, 0x38}) // size
	copy(mvhdBox[4:8], []byte{'m', 'v', 'h', 'd'})    // type
	binary.BigEndian.PutUint32(mvhdBox[12:16], info.Timescale)
	binary.BigEndian.PutUint32(mvhdBox[16:20], uint32(info.Duration.Seconds()*float64(info.Timescale)))
	
	if _, err := w.Write(mvhdBox); err != nil {
		return fmt.Errorf("failed to write mvhd box: %w", err)
	}

	return nil
}

// GetFormat returns the container format
func (mp4 *MP4Container) GetFormat() ContainerFormat {
	return FormatMP4
}

// DetectFormat detects the container format from data
func DetectFormat(data []byte) ContainerFormat {
	if len(data) < 12 {
		return FormatUnknown
	}

	// Check for MP4 ftyp box
	if bytes.Equal(data[4:8], []byte("ftyp")) {
		return FormatMP4
	}

	// Check for WebM EBML header
	if bytes.HasPrefix(data, []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return FormatWebM
	}

	// Check for MPEG-TS sync byte
	if data[0] == 0x47 {
		return FormatTS
	}

	return FormatUnknown
}

// NewContainer creates a new container parser based on format
func NewContainer(format ContainerFormat) MediaContainer {
	switch format {
	case FormatMP4, FormatFMP4:
		return NewMP4Container()
	default:
		return NewMP4Container() // Default to MP4
	}
}