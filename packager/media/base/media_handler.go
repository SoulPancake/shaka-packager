// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package base

import (
	"fmt"
	"sync"

	"github.com/SoulPancake/shaka-packager/packager/status"
)

// StreamDataType represents the type of stream data
type StreamDataType int

const (
	StreamDataUnknown StreamDataType = iota
	StreamDataStreamInfo
	StreamDataMediaSample
	StreamDataTextSample
	StreamDataSegmentInfo
	StreamDataScte35Event
	StreamDataCueEvent
)

// StreamDataTypeToString converts stream data type to string
func StreamDataTypeToString(dataType StreamDataType) string {
	switch dataType {
	case StreamDataStreamInfo:
		return "StreamInfo"
	case StreamDataMediaSample:
		return "MediaSample"
	case StreamDataTextSample:
		return "TextSample"
	case StreamDataSegmentInfo:
		return "SegmentInfo"
	case StreamDataScte35Event:
		return "Scte35Event"
	case StreamDataCueEvent:
		return "CueEvent"
	default:
		return "Unknown"
	}
}

// CueEventType represents the type of cue event
type CueEventType int

const (
	CueEventCueIn CueEventType = iota
	CueEventCueOut
	CueEventCuePoint
)

// Scte35Event represents cuepoint markers in input streams
type Scte35Event struct {
	Id                   string
	Type                 int     // Segmentation type id from SCTE35
	StartTimeInSeconds   float64
	DurationInSeconds    float64
	CueData              string
}

// CueEvent represents consolidated cue events for ad insertion
type CueEvent struct {
	Type           CueEventType
	TimeInSeconds  float64
	CueData        string
}

// SegmentInfo represents information about a media segment
type SegmentInfo struct {
	IsSubsegment           bool
	IsChunk                bool
	IsFinalChunkInSeg      bool
	IsEncrypted            bool
	StartTimestamp         int64
	Duration               int64
	SegmentNumber          int64
	KeyRotationEncryptionConfig *EncryptionConfig
}

// TextSample represents a text/subtitle sample
type TextSample struct {
	Id                string
	StartTime         int64
	Duration          int64
	Settings          map[string]string
	Body              string
	SubStreamIndex    int32
}

// StreamData holds various types of stream data
type StreamData struct {
	StreamIndex     int
	StreamDataType  StreamDataType
	StreamInfo      *StreamInfo
	MediaSample     *MediaSample
	TextSample      *TextSample
	SegmentInfo     *SegmentInfo
	Scte35Event     *Scte35Event
	CueEvent        *CueEvent
}

// Factory methods for StreamData

// FromStreamInfo creates StreamData from StreamInfo
func FromStreamInfo(streamIndex int, streamInfo *StreamInfo) *StreamData {
	return &StreamData{
		StreamIndex:    streamIndex,
		StreamDataType: StreamDataStreamInfo,
		StreamInfo:     streamInfo,
	}
}

// FromMediaSample creates StreamData from MediaSample
func FromMediaSample(streamIndex int, mediaSample *MediaSample) *StreamData {
	return &StreamData{
		StreamIndex:    streamIndex,
		StreamDataType: StreamDataMediaSample,
		MediaSample:    mediaSample,
	}
}

// FromTextSample creates StreamData from TextSample
func FromTextSample(streamIndex int, textSample *TextSample) *StreamData {
	return &StreamData{
		StreamIndex:    streamIndex,
		StreamDataType: StreamDataTextSample,
		TextSample:     textSample,
	}
}

// FromSegmentInfo creates StreamData from SegmentInfo
func FromSegmentInfo(streamIndex int, segmentInfo *SegmentInfo) *StreamData {
	return &StreamData{
		StreamIndex:    streamIndex,
		StreamDataType: StreamDataSegmentInfo,
		SegmentInfo:    segmentInfo,
	}
}

// FromScte35Event creates StreamData from Scte35Event
func FromScte35Event(streamIndex int, scte35Event *Scte35Event) *StreamData {
	return &StreamData{
		StreamIndex:    streamIndex,
		StreamDataType: StreamDataScte35Event,
		Scte35Event:    scte35Event,
	}
}

// FromCueEvent creates StreamData from CueEvent
func FromCueEvent(streamIndex int, cueEvent *CueEvent) *StreamData {
	return &StreamData{
		StreamIndex:    streamIndex,
		StreamDataType: StreamDataCueEvent,
		CueEvent:       cueEvent,
	}
}

// MediaHandler is the base interface for media processing handlers
type MediaHandler interface {
	// Initialize initializes the handler
	Initialize() *status.Status
	
	// Process processes stream data
	Process(streamData *StreamData) *status.Status
	
	// Flush flushes any pending data
	Flush() *status.Status
	
	// SetHandler sets the next handler in the chain
	SetHandler(streamIndex int, handler MediaHandler) *status.Status
	
	// SetNext sets the next handler (deprecated, use SetHandler)
	SetNext(handler MediaHandler) *status.Status
}

// BaseMediaHandler provides a base implementation of MediaHandler
type BaseMediaHandler struct {
	mu               sync.RWMutex
	nextHandlers     map[int]MediaHandler
	numInputStreams  int
	numOutputStreams int
	initialized      bool
}

// NewBaseMediaHandler creates a new BaseMediaHandler
func NewBaseMediaHandler() *BaseMediaHandler {
	return &BaseMediaHandler{
		nextHandlers: make(map[int]MediaHandler),
	}
}

// Initialize initializes the base handler
func (bmh *BaseMediaHandler) Initialize() *status.Status {
	bmh.mu.Lock()
	defer bmh.mu.Unlock()
	
	bmh.initialized = true
	return status.NewOkStatus()
}

// Process processes stream data (base implementation)
func (bmh *BaseMediaHandler) Process(streamData *StreamData) *status.Status {
	bmh.mu.RLock()
	defer bmh.mu.RUnlock()
	
	if !bmh.initialized {
		return status.NewStatus(status.PreconditionFailed, "handler not initialized")
	}
	
	// Forward to next handler
	return bmh.Dispatch(streamData)
}

// Flush flushes any pending data (base implementation)
func (bmh *BaseMediaHandler) Flush() *status.Status {
	bmh.mu.RLock()
	defer bmh.mu.RUnlock()
	
	// Flush all next handlers
	for _, handler := range bmh.nextHandlers {
		if stat := handler.Flush(); !stat.Ok() {
			return stat
		}
	}
	
	return status.NewOkStatus()
}

// SetHandler sets the next handler for a specific stream
func (bmh *BaseMediaHandler) SetHandler(streamIndex int, handler MediaHandler) *status.Status {
	bmh.mu.Lock()
	defer bmh.mu.Unlock()
	
	if handler == nil {
		return status.NewStatus(status.InvalidArgument, "handler cannot be nil")
	}
	
	bmh.nextHandlers[streamIndex] = handler
	return status.NewOkStatus()
}

// SetNext sets the next handler (for single stream)
func (bmh *BaseMediaHandler) SetNext(handler MediaHandler) *status.Status {
	return bmh.SetHandler(0, handler)
}

// Dispatch dispatches stream data to the appropriate next handler
func (bmh *BaseMediaHandler) Dispatch(streamData *StreamData) *status.Status {
	if streamData == nil {
		return status.NewStatus(status.InvalidArgument, "stream data cannot be nil")
	}
	
	handler, exists := bmh.nextHandlers[streamData.StreamIndex]
	if !exists {
		// No handler for this stream index, try default (index 0)
		if handler, exists = bmh.nextHandlers[0]; !exists {
			return status.NewStatus(status.InvalidArgument, 
				fmt.Sprintf("no handler found for stream index %d", streamData.StreamIndex))
		}
	}
	
	return handler.Process(streamData)
}

// GetNumInputStreams returns the number of input streams
func (bmh *BaseMediaHandler) GetNumInputStreams() int {
	bmh.mu.RLock()
	defer bmh.mu.RUnlock()
	return bmh.numInputStreams
}

// SetNumInputStreams sets the number of input streams
func (bmh *BaseMediaHandler) SetNumInputStreams(numStreams int) {
	bmh.mu.Lock()
	defer bmh.mu.Unlock()
	bmh.numInputStreams = numStreams
}

// GetNumOutputStreams returns the number of output streams
func (bmh *BaseMediaHandler) GetNumOutputStreams() int {
	bmh.mu.RLock()
	defer bmh.mu.RUnlock()
	return bmh.numOutputStreams
}

// SetNumOutputStreams sets the number of output streams
func (bmh *BaseMediaHandler) SetNumOutputStreams(numStreams int) {
	bmh.mu.Lock()
	defer bmh.mu.Unlock()
	bmh.numOutputStreams = numStreams
}

// IsInitialized returns whether the handler has been initialized
func (bmh *BaseMediaHandler) IsInitialized() bool {
	bmh.mu.RLock()
	defer bmh.mu.RUnlock()
	return bmh.initialized
}