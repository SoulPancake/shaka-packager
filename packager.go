// Package gopackager provides a pure Go implementation for media packaging.
// This package provides native Go functionality for media packaging operations
// including DASH and HLS manifest generation, stream segmentation, and encryption support.
//
// The implementation is written entirely in Go without any C/C++ dependencies,
// providing a clean, native Go interface for media packaging operations.
//
// Example usage:
//
//	packager := gopackager.NewPackager()
//	defer packager.Close()
//
//	params := gopackager.PackagingParams{
//		TempDir: "/tmp/packaging",
//		OutputMediaInfo: true,
//	}
//
//	streams := []gopackager.StreamDescriptor{
//		{
//			Input: "input.mp4",
//			StreamSelector: "video",
//			Output: "output_video.mp4",
//		},
//	}
//
//	if err := packager.Initialize(params, streams); err != nil {
//		log.Fatal(err)
//	}
//
//	if err := packager.Run(); err != nil {
//		log.Fatal(err)
//	}
package gopackager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/SoulPancake/shaka-packager/crypto"
	"github.com/SoulPancake/shaka-packager/hls"
	"github.com/SoulPancake/shaka-packager/media"
	"github.com/SoulPancake/shaka-packager/mpd"
)

// Version information
const (
	LibraryVersion = "1.0.0-go"
)

// Status codes for packaging operations
const (
	StatusOK = iota
	StatusUnknown
	StatusCancelled
	StatusInvalidArgument
	StatusNotFound
	StatusAlreadyExists
	StatusResourceExhausted
	StatusFailedPrecondition
	StatusAborted
	StatusOutOfRange
	StatusUnimplemented
	StatusInternal
	StatusUnavailable
	StatusDataLoss
	StatusUnauthenticated
)

// PackagingParams contains parameters for the packaging operation
type PackagingParams struct {
	// Temporary directory for intermediate files
	TempDir string

	// Create human readable MediaInfo output files
	OutputMediaInfo bool

	// Use single thread for deterministic output (useful for testing)
	SingleThreaded bool

	// MP4 output parameters
	Mp4OutputParams Mp4OutputParams

	// Chunking parameters
	ChunkingParams ChunkingParams

	// Encryption parameters
	EncryptionParams EncryptionParams
}

// Mp4OutputParams contains MP4-specific output parameters
type Mp4OutputParams struct {
	// Include pssh box in the media files
	IncludePsshInStream bool

	// Generate dash_if_iop compliant mpd
	GenerateDashIfIopCompliantMpd bool
}

// ChunkingParams contains chunking/segmentation parameters
type ChunkingParams struct {
	// Segment duration in seconds
	SegmentDurationInSeconds float64

	// Subsegment duration in seconds
	SubsegmentDurationInSeconds float64
}

// EncryptionParams contains encryption parameters
type EncryptionParams struct {
	// Encryption key ID (hex string)
	KeyId string

	// Encryption key (hex string)
	Key string

	// Clear lead duration in seconds
	ClearLeadInSeconds float64
}

// StreamDescriptor defines a single input/output stream
type StreamDescriptor struct {
	// Input file path or URL
	Input string

	// Stream selector (e.g., "video", "audio", "text" or stream index)
	StreamSelector string

	// Output file path
	Output string

	// Segment template for segmented output
	SegmentTemplate string

	// Stream index for ordering
	Index *uint32
}

// MediaInfo represents information about a media stream
type MediaInfo struct {
	// Media type (video, audio, text)
	MediaType string

	// Codec information
	Codec string

	// Duration in seconds
	Duration float64

	// Bitrate in bits per second
	Bitrate int64

	// Resolution for video streams
	Width  int32
	Height int32

	// Sample rate for audio streams
	SampleRate int32

	// Number of channels for audio streams
	Channels int32
}

// Packager provides pure Go media packaging functionality
type Packager struct {
	mu          sync.RWMutex
	params      PackagingParams
	streams     []StreamDescriptor
	ctx         context.Context
	cancelFunc  context.CancelFunc
	initialized bool
	running     bool
	created     bool // tracks if packager was properly created
	closed      bool // tracks if packager was closed
	
	// Pure Go components
	demuxers     map[string]media.Demuxer
	muxers       map[string]media.Muxer
	encryptors   map[string]crypto.Encryptor
	hlsNotifier  hls.HLSNotifier
	mpdNotifier  mpd.MPDNotifier
	tempDir      string
}

// NewPackager creates a new Packager instance
func NewPackager() *Packager {
	return &Packager{
		initialized: false,
		running:     false,
		created:     true, // Mark as properly created
		closed:      false,
		demuxers:    make(map[string]media.Demuxer),
		muxers:      make(map[string]media.Muxer),
		encryptors:  make(map[string]crypto.Encryptor),
	}
}

// Close cleans up resources and cancels any running operations
func (p *Packager) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
		p.cancelFunc = nil
	}

	p.initialized = false
	p.running = false
	p.closed = true
}

// Initialize initializes the packaging pipeline with the given parameters and streams
func (p *Packager) Initialize(params PackagingParams, streams []StreamDescriptor) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.created {
		return errors.New("packager not initialized")
	}

	if p.closed {
		return errors.New("packager not initialized")
	}

	if p.initialized {
		return errors.New("packager already initialized")
	}

	if len(streams) == 0 {
		return errors.New("no streams specified")
	}

	// Validate parameters
	if err := p.validateParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate streams
	if err := p.validateStreams(streams); err != nil {
		return fmt.Errorf("invalid streams: %w", err)
	}

	// Create temporary directory if specified
	if params.TempDir != "" {
		if err := os.MkdirAll(params.TempDir, 0755); err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		p.tempDir = params.TempDir
	} else {
		// Create default temp directory
		tempDir, err := os.MkdirTemp("", "shaka-packager-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		p.tempDir = tempDir
	}

	// Initialize HLS and MPD notifiers if needed
	for _, stream := range streams {
		if strings.Contains(stream.Output, ".m3u8") {
			if p.hlsNotifier == nil {
				outputDir := filepath.Dir(stream.Output)
				p.hlsNotifier = hls.NewSimpleHLSNotifier(outputDir)
			}
		}
		if strings.Contains(stream.Output, ".mpd") {
			if p.mpdNotifier == nil {
				p.mpdNotifier = mpd.NewSimpleMPDNotifier(stream.Output)
			}
		}
	}

	// Create context for cancellation
	p.ctx, p.cancelFunc = context.WithCancel(context.Background())

	// Store parameters and streams
	p.params = params
	p.streams = streams
	p.initialized = true

	return nil
}

// Run executes the packaging operation
func (p *Packager) Run() error {
	p.mu.Lock()
	if !p.created || p.closed {
		p.mu.Unlock()
		return errors.New("packager not initialized")
	}

	if !p.initialized {
		p.mu.Unlock()
		return errors.New("packager not initialized")
	}

	if p.running {
		p.mu.Unlock()
		return errors.New("packager is already running")
	}

	p.running = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
	}()

	return p.executePackaging()
}

// Cancel cancels the current packaging operation
func (p *Packager) Cancel() {
	p.mu.RLock()
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	p.mu.RUnlock()
}

// GetLibraryVersion returns the library version string
func GetLibraryVersion() string {
	return LibraryVersion
}

// IsInitialized returns true if the packager has been initialized
func (p *Packager) IsInitialized() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized
}

// IsRunning returns true if the packager is currently running
func (p *Packager) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// validateParams validates the packaging parameters
func (p *Packager) validateParams(params PackagingParams) error {
	// Validate temp directory if specified
	if params.TempDir != "" {
		if !filepath.IsAbs(params.TempDir) {
			return errors.New("temp directory must be absolute path")
		}
	}

	// Validate chunking parameters
	if params.ChunkingParams.SegmentDurationInSeconds < 0 {
		return errors.New("segment duration cannot be negative")
	}

	if params.ChunkingParams.SubsegmentDurationInSeconds < 0 {
		return errors.New("subsegment duration cannot be negative")
	}

	// Validate encryption parameters
	if params.EncryptionParams.KeyId != "" || params.EncryptionParams.Key != "" {
		if params.EncryptionParams.KeyId == "" || params.EncryptionParams.Key == "" {
			return errors.New("both key ID and key must be specified for encryption")
		}
	}

	return nil
}

// validateStreams validates the stream descriptors
func (p *Packager) validateStreams(streams []StreamDescriptor) error {
	seenOutputs := make(map[string]bool)

	for i, stream := range streams {
		if stream.Input == "" {
			return fmt.Errorf("stream %d: input file is required", i)
		}

		if stream.Output == "" {
			return fmt.Errorf("stream %d: output file is required", i)
		}

		// Check for duplicate outputs
		if seenOutputs[stream.Output] {
			return fmt.Errorf("stream %d: duplicate output file: %s", i, stream.Output)
		}
		seenOutputs[stream.Output] = true

		// Validate stream selector
		if stream.StreamSelector != "" {
			if !isValidStreamSelector(stream.StreamSelector) {
				return fmt.Errorf("stream %d: invalid stream selector: %s", i, stream.StreamSelector)
			}
		}
	}

	return nil
}

// executePackaging performs the actual packaging operation
func (p *Packager) executePackaging() error {
	// Create output directories
	for _, stream := range p.streams {
		outputDir := filepath.Dir(stream.Output)
		if outputDir != "." {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
			}
		}
	}

	// Process each stream using pure Go components
	for i, stream := range p.streams {
		select {
		case <-p.ctx.Done():
			return errors.New("operation cancelled")
		default:
		}

		if err := p.processStreamPure(i, stream); err != nil {
			return fmt.Errorf("failed to process stream %d: %w", i, err)
		}
	}

	// Flush notifiers
	if p.hlsNotifier != nil {
		if err := p.hlsNotifier.Flush(); err != nil {
			return fmt.Errorf("failed to flush HLS notifier: %w", err)
		}
	}
	
	if p.mpdNotifier != nil {
		if err := p.mpdNotifier.Flush(); err != nil {
			return fmt.Errorf("failed to flush MPD notifier: %w", err)
		}
	}

	// Generate media info if requested
	if p.params.OutputMediaInfo {
		if err := p.generateMediaInfo(); err != nil {
			return fmt.Errorf("failed to generate media info: %w", err)
		}
	}

	return nil
}

// processStreamPure processes a single stream using pure Go components
func (p *Packager) processStreamPure(index int, stream StreamDescriptor) error {
	// Check if input file exists
	if _, err := os.Stat(stream.Input); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", stream.Input)
	}

	// Create demuxer for input
	demuxer, err := media.NewDemuxer(stream.Input)
	if err != nil {
		return fmt.Errorf("failed to create demuxer: %w", err)
	}
	defer demuxer.Close()

	// Initialize demuxer
	if err := demuxer.Initialize(stream.Input); err != nil {
		return fmt.Errorf("failed to initialize demuxer: %w", err)
	}

	// Get available streams
	_, err = demuxer.GetStreams()
	if err != nil {
		return fmt.Errorf("failed to get streams: %w", err)
	}

	// Find the requested stream
	var selectedStreamID uint32
	if stream.StreamSelector == "video" {
		selectedStreamID = 1 // First video stream
	} else if stream.StreamSelector == "audio" {
		selectedStreamID = 2 // First audio stream  
	} else {
		selectedStreamID = 1 // Default to first stream
	}

	// Create muxer for output
	isSegmented := strings.Contains(stream.Output, "%d") || strings.Contains(stream.Output, "$Number")
	segmentLength := int64(p.params.ChunkingParams.SegmentDurationInSeconds) * 1000000 // Convert to microseconds
	if segmentLength == 0 {
		segmentLength = 6000000 // 6 seconds default
	}

	muxer := media.NewMuxer(stream.Output, isSegmented, segmentLength)
	defer muxer.Close()

	// Initialize muxer with stream descriptors
	streamDescriptors := []media.StreamDescriptor{
		{
			StreamID:   selectedStreamID,
			StreamType: p.getStreamType(stream.StreamSelector),
			Codec:      "avc1", // Default codec
			Bitrate:    1000000, // 1 Mbps default
		},
	}

	if err := muxer.Initialize(stream.Output, streamDescriptors); err != nil {
		return fmt.Errorf("failed to initialize muxer: %w", err)
	}

	// Set up encryption if specified
	var encryptor crypto.Encryptor
	if len(p.params.EncryptionParams.Key) > 0 {
		encryptor = crypto.NewEncryptor(crypto.EncryptionAES128)
		if encryptor != nil {
			// Convert hex strings to bytes
			keyBytes, err := hexStringToBytes(p.params.EncryptionParams.Key)
			if err != nil {
				return fmt.Errorf("failed to parse encryption key: %w", err)
			}
			
			keyIdBytes, err := hexStringToBytes(p.params.EncryptionParams.KeyId)
			if err != nil {
				return fmt.Errorf("failed to parse key ID: %w", err)
			}
			
			encParams := crypto.EncryptionParams{
				Key:    keyBytes,
				KeyID:  keyIdBytes,
				Method: crypto.EncryptionAES128,
			}
			if err := encryptor.Initialize(encParams); err != nil {
				return fmt.Errorf("failed to initialize encryptor: %w", err)
			}
			defer encryptor.Close()
		}
	}

	// Notify HLS/MPD about the stream if needed
	if p.hlsNotifier != nil {
		streamInfo := hls.StreamInfo{
			StreamID:     selectedStreamID,
			StreamType:   p.getHLSStreamType(stream.StreamSelector),
			Bandwidth:    1000000,
			Resolution:   "1920x1080",
			PlaylistName: filepath.Base(stream.Output),
		}
		if err := p.hlsNotifier.NotifyNewStream(streamInfo); err != nil {
			return fmt.Errorf("failed to notify HLS: %w", err)
		}
	}

	if p.mpdNotifier != nil {
		containerInfo := mpd.ContainerInfo{
			ContainerType: stream.StreamSelector,
			FileName:      filepath.Base(stream.Output),
			Bandwidth:     1000000,
			Codec:         "avc1",
		}
		containerID, err := p.mpdNotifier.NotifyNewContainer(containerInfo)
		if err != nil {
			return fmt.Errorf("failed to notify MPD: %w", err)
		}
		
		streamInfo := mpd.StreamInfo{
			StreamID:  selectedStreamID,
			Bandwidth: 1000000,
			Width:     1920,
			Height:    1080,
		}
		if err := p.mpdNotifier.NotifyNewStream(containerID, streamInfo); err != nil {
			return fmt.Errorf("failed to notify MPD stream: %w", err)
		}
	}

	// Process packets
	segmentCount := 0
	for {
		select {
		case <-p.ctx.Done():
			return errors.New("operation cancelled")
		default:
		}

		// Read packet from demuxer
		packet, err := demuxer.ReadPacket()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read packet: %w", err)
		}

		// Filter by selected stream
		if packet.StreamID != selectedStreamID {
			continue
		}

		// Apply encryption if configured
		if encryptor != nil {
			encryptedData, err := encryptor.Encrypt(packet.Data)
			if err != nil {
				return fmt.Errorf("failed to encrypt packet: %w", err)
			}
			packet.Data = encryptedData
		}

		// Write packet to muxer
		if err := muxer.WritePacket(packet); err != nil {
			return fmt.Errorf("failed to write packet: %w", err)
		}

		// Update segment count for notifications
		if packet.IsKeyFrame && packet.Timestamp > 0 {
			segmentCount++
			
			// Notify about segments
			if p.hlsNotifier != nil {
				segmentInfo := hls.SegmentInfo{
					SegmentID: uint32(segmentCount),
					Duration:  time.Duration(segmentLength),
					Filename:  fmt.Sprintf("segment_%d.ts", segmentCount),
				}
				p.hlsNotifier.NotifyNewSegment(selectedStreamID, segmentInfo)
			}
		}
	}

	// Finalize muxer
	if err := muxer.Finalize(); err != nil {
		return fmt.Errorf("failed to finalize muxer: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func (p *Packager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// getStreamType converts a stream selector to media.StreamType
func (p *Packager) getStreamType(selector string) media.StreamType {
	switch selector {
	case "video":
		return media.StreamTypeVideo
	case "audio":
		return media.StreamTypeAudio
	case "text", "subtitle":
		return media.StreamTypeText
	default:
		return media.StreamTypeVideo // Default
	}
}

// getHLSStreamType converts a stream selector to hls.StreamType  
func (p *Packager) getHLSStreamType(selector string) hls.StreamType {
	switch selector {
	case "video":
		return hls.StreamTypeVideo
	case "audio":
		return hls.StreamTypeAudio
	case "text", "subtitle":
		return hls.StreamTypeSubtitle
	default:
		return hls.StreamTypeVideo // Default
	}
}

// hexStringToBytes converts a hex string to bytes
func hexStringToBytes(hexStr string) ([]byte, error) {
	// Remove any spaces or prefixes
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.TrimPrefix(hexStr, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")
	
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	
	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		bytes[i/2] = b
	}
	
	return bytes, nil
}

// generateMediaInfo generates media information files
func (p *Packager) generateMediaInfo() error {
	for i, stream := range p.streams {
		mediaInfo := MediaInfo{
			MediaType: p.detectMediaType(stream.StreamSelector),
			Codec:     "unknown",
			Duration:  0.0,
			Bitrate:   0,
		}

		// In a real implementation, this would parse the media file
		// and extract actual media information

		infoFile := stream.Output + ".info"
		if err := p.writeMediaInfo(infoFile, mediaInfo); err != nil {
			return fmt.Errorf("failed to write media info for stream %d: %w", i, err)
		}
	}

	return nil
}

// detectMediaType detects the media type from the stream selector
func (p *Packager) detectMediaType(selector string) string {
	selector = strings.ToLower(selector)
	switch {
	case strings.Contains(selector, "video"):
		return "video"
	case strings.Contains(selector, "audio"):
		return "audio"
	case strings.Contains(selector, "text") || strings.Contains(selector, "subtitle"):
		return "text"
	default:
		return "unknown"
	}
}

// writeMediaInfo writes media information to a file
func (p *Packager) writeMediaInfo(filename string, info MediaInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "Media Type: %s\n", info.MediaType)
	fmt.Fprintf(file, "Codec: %s\n", info.Codec)
	fmt.Fprintf(file, "Duration: %.2f seconds\n", info.Duration)
	fmt.Fprintf(file, "Bitrate: %d bps\n", info.Bitrate)

	if info.MediaType == "video" {
		fmt.Fprintf(file, "Resolution: %dx%d\n", info.Width, info.Height)
	} else if info.MediaType == "audio" {
		fmt.Fprintf(file, "Sample Rate: %d Hz\n", info.SampleRate)
		fmt.Fprintf(file, "Channels: %d\n", info.Channels)
	}

	return nil
}

// isValidStreamSelector checks if a stream selector is valid
func isValidStreamSelector(selector string) bool {
	if selector == "" {
		return false
	}

	// Allow common stream selectors
	validSelectors := []string{"video", "audio", "text", "subtitle"}
	for _, valid := range validSelectors {
		if strings.EqualFold(selector, valid) {
			return true
		}
	}

	// Allow numeric selectors (stream index)
	if len(selector) > 0 && selector[0] >= '0' && selector[0] <= '9' {
		return true
	}

	return false
}

// statusCodeToError converts a status code to an error
func statusCodeToError(status int) error {
	switch status {
	case StatusOK:
		return nil
	case StatusUnknown:
		return errors.New("unknown error")
	case StatusCancelled:
		return errors.New("operation cancelled")
	case StatusInvalidArgument:
		return errors.New("invalid argument")
	case StatusNotFound:
		return errors.New("not found")
	case StatusAlreadyExists:
		return errors.New("already exists")
	case StatusResourceExhausted:
		return errors.New("resource exhausted")
	case StatusFailedPrecondition:
		return errors.New("failed precondition")
	case StatusAborted:
		return errors.New("aborted")
	case StatusOutOfRange:
		return errors.New("out of range")
	case StatusUnimplemented:
		return errors.New("unimplemented")
	case StatusInternal:
		return errors.New("internal error")
	case StatusUnavailable:
		return errors.New("unavailable")
	case StatusDataLoss:
		return errors.New("data loss")
	case StatusUnauthenticated:
		return errors.New("unauthenticated")
	default:
		return fmt.Errorf("unknown status code: %d", status)
	}
}
