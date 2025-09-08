// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// AdCueGeneratorParams contains parameters for ad cue generation.
type AdCueGeneratorParams struct {
	// Path to the cue points file for ad insertion. The file contains cue points
	// in seconds, one per line.
	CuePointsFile string
}

// NewAdCueGeneratorParams creates an AdCueGeneratorParams with default values.
func NewAdCueGeneratorParams() AdCueGeneratorParams {
	return AdCueGeneratorParams{}
}