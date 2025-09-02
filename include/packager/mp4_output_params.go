// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// Mp4OutputParams represents MP4 (ISO-BMFF) output related parameters.
type Mp4OutputParams struct {
	// Include PSSH in the encrypted stream. CMAF and DASH-IF recommends carrying
	// license acquisition information in the manifest and not duplicate the
	// information in the stream. (This is not a hard requirement so we are still
	// CMAF compatible even if PSSH is included in the stream.)
	IncludePSSHInStream bool

	// Indicates whether a 'sidx' box should be generated in the media segments.
	// Note that it is required by spec if segment_template contains $Times$ specifier.
	GenerateSIDXInMediaSegments bool

	// Enable LL-DASH streaming.
	// Each segment consists of many fragments, and each fragment contains one
	// chunk. A chunk is the smallest unit and is constructed of a single moof
	// and mdat atom. Each chunk is uploaded immediately upon creation,
	// decoupling latency from segment duration.
	LowLatencyDashMode bool
}

// NewMp4OutputParams creates a Mp4OutputParams with default values.
func NewMp4OutputParams() Mp4OutputParams {
	return Mp4OutputParams{
		IncludePSSHInStream:         true,
		GenerateSIDXInMediaSegments: true,
		LowLatencyDashMode:          false,
	}
}