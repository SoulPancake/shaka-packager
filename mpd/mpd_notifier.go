// Package mpd provides DASH MPD (Media Presentation Description) functionality
package mpd

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// MPDNotifier handles DASH MPD generation and updates
type MPDNotifier interface {
	NotifyNewContainer(containerInfo ContainerInfo) (uint32, error)
	NotifyNewStream(containerID uint32, streamInfo StreamInfo) error  
	NotifyNewSegment(containerID uint32, segmentInfo SegmentInfo) error
	NotifyEncryptionUpdate(containerID uint32, drmInfo DRMInfo) error
	Flush() error
	Close() error
}

// ContainerInfo contains information about a media container
type ContainerInfo struct {
	ContainerType string
	FileName      string
	Bandwidth     uint64
	Language      string
	Codec         string
}

// StreamInfo contains information about a media stream  
type StreamInfo struct {
	StreamID     uint32
	StreamType   StreamType
	Codec        string
	Bandwidth    uint64
	Width        uint32
	Height       uint32
	FrameRate    float64
	SampleRate   uint32
	Channels     uint32
	Language     string
}

// StreamType represents the type of stream
type StreamType int

const (
	StreamTypeUnknown StreamType = iota
	StreamTypeVideo
	StreamTypeAudio
	StreamTypeText
)

// SegmentInfo contains information about a media segment
type SegmentInfo struct {
	SegmentID  uint32
	Filename   string
	Duration   time.Duration
	Size       uint64
	Timestamp  int64
	IsKeyFrame bool
}

// DRMInfo contains DRM/encryption information
type DRMInfo struct {
	SchemeUUID     string
	SystemID       []byte
	PSSH           []byte
	KeyIDs         [][]byte
	ProtectionInfo string
}

// SimpleMPDNotifier implements basic DASH MPD generation
type SimpleMPDNotifier struct {
	outputPath    string
	mpd          *MPD
	containers   map[uint32]*Container
	nextID       uint32
}

// NewSimpleMPDNotifier creates a new simple MPD notifier
func NewSimpleMPDNotifier(outputPath string) *SimpleMPDNotifier {
	return &SimpleMPDNotifier{
		outputPath: outputPath,
		mpd:        NewMPD(),
		containers: make(map[uint32]*Container),
		nextID:     1,
	}
}

// NotifyNewContainer adds a new container to the MPD
func (m *SimpleMPDNotifier) NotifyNewContainer(containerInfo ContainerInfo) (uint32, error) {
	containerID := m.nextID
	m.nextID++
	
	container := &Container{
		ID:            containerID,
		ContainerType: containerInfo.ContainerType,
		FileName:      containerInfo.FileName,
		Bandwidth:     containerInfo.Bandwidth,
		Language:      containerInfo.Language,
		Codec:         containerInfo.Codec,
		Streams:       make([]*Stream, 0),
		Segments:      make([]SegmentInfo, 0),
	}
	
	m.containers[containerID] = container
	
	// Add to appropriate adaptation set based on type
	if containerInfo.ContainerType == "video" {
		m.mpd.AddVideoContainer(container)
	} else if containerInfo.ContainerType == "audio" {
		m.mpd.AddAudioContainer(container)  
	} else if containerInfo.ContainerType == "text" {
		m.mpd.AddTextContainer(container)
	}
	
	return containerID, nil
}

// NotifyNewStream adds a new stream to a container
func (m *SimpleMPDNotifier) NotifyNewStream(containerID uint32, streamInfo StreamInfo) error {
	container, exists := m.containers[containerID]
	if !exists {
		return fmt.Errorf("container %d not found", containerID)
	}
	
	stream := &Stream{
		ID:         streamInfo.StreamID,
		StreamType: streamInfo.StreamType,
		Codec:      streamInfo.Codec,
		Bandwidth:  streamInfo.Bandwidth,
		Width:      streamInfo.Width,
		Height:     streamInfo.Height,
		FrameRate:  streamInfo.FrameRate,
		SampleRate: streamInfo.SampleRate,
		Channels:   streamInfo.Channels,
		Language:   streamInfo.Language,
	}
	
	container.Streams = append(container.Streams, stream)
	
	return nil
}

// NotifyNewSegment adds a new segment to a container
func (m *SimpleMPDNotifier) NotifyNewSegment(containerID uint32, segmentInfo SegmentInfo) error {
	container, exists := m.containers[containerID]
	if !exists {
		return fmt.Errorf("container %d not found", containerID)
	}
	
	container.Segments = append(container.Segments, segmentInfo)
	return nil
}

// NotifyEncryptionUpdate updates encryption information for a container
func (m *SimpleMPDNotifier) NotifyEncryptionUpdate(containerID uint32, drmInfo DRMInfo) error {
	container, exists := m.containers[containerID]
	if !exists {
		return fmt.Errorf("container %d not found", containerID)
	}
	
	container.DRMInfo = &drmInfo
	return nil
}

// Flush writes the MPD to file
func (m *SimpleMPDNotifier) Flush() error {
	// Update MPD timings and durations
	m.mpd.UpdateTimings()
	
	file, err := os.Create(m.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create MPD file: %w", err)
	}
	defer file.Close()
	
	return m.mpd.WriteTo(file)
}

// Close cleans up resources
func (m *SimpleMPDNotifier) Close() error {
	return m.Flush()
}

// MPD represents a DASH Media Presentation Description
type MPD struct {
	XMLName                xml.Name           `xml:"MPD"`
	Xmlns                  string            `xml:"xmlns,attr"`
	XmlnsXsi               string            `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation      string            `xml:"xsi:schemaLocation,attr"`
	Type                   string            `xml:"type,attr"`
	MediaPresentationDuration string         `xml:"mediaPresentationDuration,attr"`
	MinBufferTime          string            `xml:"minBufferTime,attr"`
	Profiles               string            `xml:"profiles,attr"`
	Periods                []Period          `xml:"Period"`
}

// Period represents a period in the MPD
type Period struct {
	XMLName         xml.Name          `xml:"Period"`
	ID              string           `xml:"id,attr"`
	Start           string           `xml:"start,attr,omitempty"`
	Duration        string           `xml:"duration,attr,omitempty"`
	AdaptationSets  []AdaptationSet  `xml:"AdaptationSet"`
}

// AdaptationSet represents an adaptation set
type AdaptationSet struct {
	XMLName          xml.Name         `xml:"AdaptationSet"`
	ID               uint32          `xml:"id,attr"`
	MimeType         string          `xml:"mimeType,attr"`
	Codecs           string          `xml:"codecs,attr"`
	Lang             string          `xml:"lang,attr,omitempty"`
	Width            uint32          `xml:"width,attr,omitempty"`
	Height           uint32          `xml:"height,attr,omitempty"`
	FrameRate        string          `xml:"frameRate,attr,omitempty"`
	AudioSamplingRate string         `xml:"audioSamplingRate,attr,omitempty"`
	Representations  []Representation `xml:"Representation"`
	ContentProtection []ContentProtection `xml:"ContentProtection"`
}

// Representation represents a media representation
type Representation struct {
	XMLName         xml.Name        `xml:"Representation"`
	ID              string         `xml:"id,attr"`
	Bandwidth       uint64         `xml:"bandwidth,attr"`
	Width           uint32         `xml:"width,attr,omitempty"`
	Height          uint32         `xml:"height,attr,omitempty"`
	FrameRate       string         `xml:"frameRate,attr,omitempty"`
	AudioSamplingRate string       `xml:"audioSamplingRate,attr,omitempty"`
	Codecs          string         `xml:"codecs,attr,omitempty"`
	BaseURL         string         `xml:"BaseURL"`
	SegmentBase     *SegmentBase   `xml:"SegmentBase,omitempty"`
	SegmentList     *SegmentList   `xml:"SegmentList,omitempty"`
}

// ContentProtection represents content protection (DRM) information
type ContentProtection struct {
	XMLName    xml.Name `xml:"ContentProtection"`
	SchemeURI  string  `xml:"schemeIdUri,attr"`
	Value      string  `xml:"value,attr,omitempty"`
	CENCPSSH   string  `xml:"cenc:pssh,omitempty"`
}

// SegmentBase represents segment base information
type SegmentBase struct {
	XMLName           xml.Name          `xml:"SegmentBase"`
	IndexRange        string           `xml:"indexRange,attr,omitempty"`
	Initialization    *Initialization  `xml:"Initialization,omitempty"`
}

// SegmentList represents a list of segments
type SegmentList struct {
	XMLName        xml.Name       `xml:"SegmentList"`
	Duration       uint64        `xml:"duration,attr,omitempty"`
	Timescale      uint32        `xml:"timescale,attr,omitempty"`
	Initialization *Initialization `xml:"Initialization,omitempty"`
	SegmentURLs    []SegmentURL   `xml:"SegmentURL"`
}

// Initialization represents initialization information
type Initialization struct {
	XMLName xml.Name `xml:"Initialization"`
	Range   string  `xml:"range,attr,omitempty"`
	SourceURL string `xml:"sourceURL,attr,omitempty"`
}

// SegmentURL represents a segment URL
type SegmentURL struct {
	XMLName xml.Name `xml:"SegmentURL"`
	Media   string  `xml:"media,attr"`
	Range   string  `xml:"mediaRange,attr,omitempty"`
}

// Container represents a media container with streams and segments
type Container struct {
	ID            uint32
	ContainerType string
	FileName      string
	Bandwidth     uint64
	Language      string
	Codec         string
	Streams       []*Stream
	Segments      []SegmentInfo
	DRMInfo       *DRMInfo
}

// Stream represents a media stream
type Stream struct {
	ID         uint32
	StreamType StreamType
	Codec      string
	Bandwidth  uint64
	Width      uint32
	Height     uint32
	FrameRate  float64
	SampleRate uint32
	Channels   uint32
	Language   string
}

// NewMPD creates a new MPD
func NewMPD() *MPD {
	mpd := &MPD{
		Xmlns:                  "urn:mpeg:dash:schema:mpd:2011",
		XmlnsXsi:               "http://www.w3.org/2001/XMLSchema-instance",
		XsiSchemaLocation:      "urn:mpeg:dash:schema:mpd:2011 DASH-MPD.xsd",
		Type:                   "static",
		MediaPresentationDuration: "PT0H0M30.000S",
		MinBufferTime:          "PT1.5S",
		Profiles:               "urn:mpeg:dash:profile:isoff-main:2011",
		Periods:                make([]Period, 0),
	}
	
	// Add default period
	period := Period{
		ID:             "0",
		AdaptationSets: make([]AdaptationSet, 0),
	}
	mpd.Periods = append(mpd.Periods, period)
	
	return mpd
}

// AddVideoContainer adds a video container to the MPD
func (mpd *MPD) AddVideoContainer(container *Container) {
	period := &mpd.Periods[0]
	
	adaptationSet := AdaptationSet{
		ID:              uint32(len(period.AdaptationSets) + 1),
		MimeType:        "video/mp4",
		Codecs:          container.Codec,
		Representations: make([]Representation, 0),
	}
	
	// Add streams as representations
	for _, stream := range container.Streams {
		representation := Representation{
			ID:        fmt.Sprintf("%d", stream.ID),
			Bandwidth: stream.Bandwidth,
			Width:     stream.Width,
			Height:    stream.Height,
			Codecs:    stream.Codec,
			BaseURL:   container.FileName,
		}
		
		if stream.FrameRate > 0 {
			representation.FrameRate = fmt.Sprintf("%.3f", stream.FrameRate)
		}
		
		adaptationSet.Representations = append(adaptationSet.Representations, representation)
	}
	
	// Add DRM information if present
	if container.DRMInfo != nil {
		contentProtection := ContentProtection{
			SchemeURI: container.DRMInfo.SchemeUUID,
		}
		adaptationSet.ContentProtection = append(adaptationSet.ContentProtection, contentProtection)
	}
	
	period.AdaptationSets = append(period.AdaptationSets, adaptationSet)
}

// AddAudioContainer adds an audio container to the MPD
func (mpd *MPD) AddAudioContainer(container *Container) {
	period := &mpd.Periods[0]
	
	adaptationSet := AdaptationSet{
		ID:              uint32(len(period.AdaptationSets) + 1),
		MimeType:        "audio/mp4",
		Codecs:          container.Codec,
		Lang:            container.Language,
		Representations: make([]Representation, 0),
	}
	
	// Add streams as representations
	for _, stream := range container.Streams {
		representation := Representation{
			ID:        fmt.Sprintf("%d", stream.ID),
			Bandwidth: stream.Bandwidth,
			Codecs:    stream.Codec,
			BaseURL:   container.FileName,
		}
		
		if stream.SampleRate > 0 {
			representation.AudioSamplingRate = fmt.Sprintf("%d", stream.SampleRate)
		}
		
		adaptationSet.Representations = append(adaptationSet.Representations, representation)
	}
	
	// Add DRM information if present
	if container.DRMInfo != nil {
		contentProtection := ContentProtection{
			SchemeURI: container.DRMInfo.SchemeUUID,
		}
		adaptationSet.ContentProtection = append(adaptationSet.ContentProtection, contentProtection)
	}
	
	period.AdaptationSets = append(period.AdaptationSets, adaptationSet)
}

// AddTextContainer adds a text container to the MPD
func (mpd *MPD) AddTextContainer(container *Container) {
	period := &mpd.Periods[0]
	
	adaptationSet := AdaptationSet{
		ID:              uint32(len(period.AdaptationSets) + 1),
		MimeType:        "text/vtt",
		Lang:            container.Language,
		Representations: make([]Representation, 0),
	}
	
	// Add streams as representations
	for _, stream := range container.Streams {
		representation := Representation{
			ID:      fmt.Sprintf("%d", stream.ID),
			BaseURL: container.FileName,
		}
		
		adaptationSet.Representations = append(adaptationSet.Representations, representation)
	}
	
	period.AdaptationSets = append(period.AdaptationSets, adaptationSet)
}

// UpdateTimings updates the timing information in the MPD
func (mpd *MPD) UpdateTimings() {
	// Calculate total duration from all containers
	// This is a simplified implementation
	mpd.MediaPresentationDuration = "PT0H1M0.000S" // 1 minute default
}

// WriteTo writes the MPD to a writer
func (mpd *MPD) WriteTo(w io.Writer) error {
	// Write XML header
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	
	// Sort adaptation sets by ID for consistency
	for i := range mpd.Periods {
		sort.Slice(mpd.Periods[i].AdaptationSets, func(j, k int) bool {
			return mpd.Periods[i].AdaptationSets[j].ID < mpd.Periods[i].AdaptationSets[k].ID
		})
	}
	
	// Marshal and write MPD
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	return encoder.Encode(mpd)
}