// Package gopackager provides a Go wrapper around the shaka-packager C++ library.
// This package demonstrates how Go can interface with the existing C++ codebase
// while maintaining minimal changes to the original implementation.
//
// The wrapper provides access to core packaging functionality including:
// - Media packaging and segmentation
// - DASH and HLS manifest generation
// - Encryption and DRM support
// - Multiple input/output stream handling
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

/*
#cgo CPPFLAGS: -I../include
#cgo LDFLAGS: -L../build -lpackager -lstdc++

#include <stdlib.h>

// Forward declarations for C++ types
typedef struct PackagerWrapper PackagerWrapper;
typedef struct PackagingParamsWrapper PackagingParamsWrapper;
typedef struct StreamDescriptorWrapper StreamDescriptorWrapper;

// C wrapper functions
extern PackagerWrapper* packager_create();
extern void packager_destroy(PackagerWrapper* p);
extern int packager_initialize(PackagerWrapper* p, PackagingParamsWrapper* params, StreamDescriptorWrapper* streams, int stream_count);
extern int packager_run(PackagerWrapper* p);
extern void packager_cancel(PackagerWrapper* p);
extern char* packager_get_library_version();

// Parameter wrapper functions
extern PackagingParamsWrapper* packaging_params_create();
extern void packaging_params_destroy(PackagingParamsWrapper* p);
extern void packaging_params_set_temp_dir(PackagingParamsWrapper* p, const char* temp_dir);
extern void packaging_params_set_output_media_info(PackagingParamsWrapper* p, int output_media_info);
extern void packaging_params_set_single_threaded(PackagingParamsWrapper* p, int single_threaded);

// Stream descriptor wrapper functions  
extern StreamDescriptorWrapper* stream_descriptor_create();
extern void stream_descriptor_destroy(StreamDescriptorWrapper* s);
extern void stream_descriptor_set_input(StreamDescriptorWrapper* s, const char* input);
extern void stream_descriptor_set_stream_selector(StreamDescriptorWrapper* s, const char* stream_selector);
extern void stream_descriptor_set_output(StreamDescriptorWrapper* s, const char* output);
extern void stream_descriptor_set_segment_template(StreamDescriptorWrapper* s, const char* segment_template);
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Status codes matching the C++ Status enum
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

// Packager wraps the C++ Packager class
type Packager struct {
	ptr *C.PackagerWrapper
}

// NewPackager creates a new Packager instance
func NewPackager() *Packager {
	ptr := C.packager_create()
	if ptr == nil {
		return nil
	}
	
	return &Packager{ptr: ptr}
}

// Close destroys the underlying C++ object and frees memory
func (p *Packager) Close() {
	if p.ptr != nil {
		C.packager_destroy(p.ptr)
		p.ptr = nil
	}
}

// Initialize initializes the packaging pipeline
func (p *Packager) Initialize(params PackagingParams, streams []StreamDescriptor) error {
	if p.ptr == nil {
		return errors.New("packager not initialized")
	}
	
	if len(streams) == 0 {
		return errors.New("no streams specified")
	}
	
	// Create parameters wrapper
	paramsWrapper := C.packaging_params_create()
	if paramsWrapper == nil {
		return errors.New("failed to create packaging parameters")
	}
	defer C.packaging_params_destroy(paramsWrapper)
	
	// Set parameters
	if params.TempDir != "" {
		cTempDir := C.CString(params.TempDir)
		defer C.free(unsafe.Pointer(cTempDir))
		C.packaging_params_set_temp_dir(paramsWrapper, cTempDir)
	}
	
	C.packaging_params_set_output_media_info(paramsWrapper, boolToInt(params.OutputMediaInfo))
	C.packaging_params_set_single_threaded(paramsWrapper, boolToInt(params.SingleThreaded))
	
	// Create stream descriptors
	streamWrappers := make([]*C.StreamDescriptorWrapper, len(streams))
	for i, stream := range streams {
		streamWrapper := C.stream_descriptor_create()
		if streamWrapper == nil {
			// Clean up already created wrappers
			for j := 0; j < i; j++ {
				C.stream_descriptor_destroy(streamWrappers[j])
			}
			return errors.New("failed to create stream descriptor")
		}
		streamWrappers[i] = streamWrapper
		
		// Set stream parameters
		if stream.Input != "" {
			cInput := C.CString(stream.Input)
			defer C.free(unsafe.Pointer(cInput))
			C.stream_descriptor_set_input(streamWrapper, cInput)
		}
		
		if stream.StreamSelector != "" {
			cSelector := C.CString(stream.StreamSelector)
			defer C.free(unsafe.Pointer(cSelector))
			C.stream_descriptor_set_stream_selector(streamWrapper, cSelector)
		}
		
		if stream.Output != "" {
			cOutput := C.CString(stream.Output)
			defer C.free(unsafe.Pointer(cOutput))
			C.stream_descriptor_set_output(streamWrapper, cOutput)
		}
		
		if stream.SegmentTemplate != "" {
			cTemplate := C.CString(stream.SegmentTemplate)
			defer C.free(unsafe.Pointer(cTemplate))
			C.stream_descriptor_set_segment_template(streamWrapper, cTemplate)
		}
	}
	
	// Clean up stream wrappers
	defer func() {
		for _, wrapper := range streamWrappers {
			C.stream_descriptor_destroy(wrapper)
		}
	}()
	
	// Call C++ initialize
	if len(streamWrappers) > 0 {
		result := C.packager_initialize(p.ptr, paramsWrapper, streamWrappers[0], C.int(len(streams)))
		return statusToError(int(result))
	} else {
		return errors.New("no streams provided")
	}
}

// Run executes the packaging pipeline to completion
func (p *Packager) Run() error {
	if p.ptr == nil {
		return errors.New("packager not initialized")
	}
	
	result := C.packager_run(p.ptr)
	return statusToError(int(result))
}

// Cancel cancels the packaging operation (can be called from another goroutine)
func (p *Packager) Cancel() {
	if p.ptr != nil {
		C.packager_cancel(p.ptr)
	}
}

// GetLibraryVersion returns the version of the underlying C++ library
func GetLibraryVersion() string {
	cVersion := C.packager_get_library_version()
	if cVersion == nil {
		return "unknown"
	}
	defer C.free(unsafe.Pointer(cVersion))
	return C.GoString(cVersion)
}

// Helper functions

func boolToInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

func statusToError(status int) error {
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