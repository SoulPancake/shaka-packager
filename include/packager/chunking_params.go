// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// ChunkingParams represents chunking (segmentation) related parameters.
type ChunkingParams struct {
	// Segment duration in seconds.
	SegmentDurationInSeconds float64
	// Subsegment duration in seconds. Should not be larger than the segment duration.
	SubsegmentDurationInSeconds float64

	// Force segments to begin with stream access points. Actual segment duration
	// may not be exactly what is specified by segment_duration.
	SegmentSAPAligned bool
	// Force subsegments to begin with stream access points. Actual subsegment
	// duration may not be exactly what is specified by subsegment_duration.
	// Setting to subsegment_sap_aligned to true but segment_sap_aligned to false
	// is not allowed.
	SubsegmentSAPAligned bool
	// Enable LL-DASH streaming.
	// Each segment consists of many fragments, and each fragment contains one
	// chunk. A chunk is the smallest unit and is constructed of a single moof
	// and mdat atom. Each chunk is uploaded immediately upon creation,
	// decoupling latency from segment duration.
	LowLatencyDashMode bool

	// Indicates the startNumber in DASH SegmentTemplate and HLS segment name.
	StartSegmentNumber int64
}

// NewChunkingParams creates a ChunkingParams with default values.
func NewChunkingParams() ChunkingParams {
	return ChunkingParams{
		SegmentDurationInSeconds:    0,
		SubsegmentDurationInSeconds: 0,
		SegmentSAPAligned:           true,
		SubsegmentSAPAligned:        true,
		LowLatencyDashMode:          false,
		StartSegmentNumber:          1,
	}
}