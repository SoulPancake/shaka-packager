// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

import "time"

// MpdParams contains DASH MPD related parameters.
type MpdParams struct {
	// Base URLs for the MPD. The values will be added as <BaseURL> elements
	// under the <MPD> element. Ignored if empty.
	BaseUrls []string
	// Indicates whether the DASH output is a static DASH or dynamic DASH content.
	// If not specified, a static DASH content is generated.
	GenerateStaticLiveMpd bool
	// Available only if output is live (dynamic DASH).
	// UTCTiming schemeUri. Typical values are
	// "urn:mpeg:dash:utc:ntp:2014" (the NTP server) or
	// "urn:mpeg:dash:utc:http-xsdate:2014" (XS Date server).
	UtcTimingSchemeUri string
	// Available only if output is live (dynamic DASH).
	// UTCTiming value.
	UtcTimingValue string
	// Available only if output is live (dynamic DASH).
	// Indicates to DASH clients how often to refresh the media presentation
	// description in seconds. This will add the minimumUpdatePeriod attribute
	// to the <MPD> element.
	MinimumUpdatePeriod float64
	// Indicates how long the DASH client can delay before refreshing the media
	// presentation description. This is the time window during which the client
	// may not refresh. This will add the timeShiftBufferDepth attribute to the
	// <MPD> element for dynamic DASH.
	TimeShiftBufferDepth float64
	// Offset with respect to the wall clock time for MPD availabilityStartTime
	// and availabilityEndTime values, in seconds. This value is used for
	// live profile only.
	SuggestedPresentationDelay float64
	// Specifies, in seconds, a common duration used in the definition of the
	// Representation data segments. This value is used for live profile only.
	MinBufferTime float64
	// Indicates the target duration for the segments (in seconds). Actual
	// segment duration may not be exactly as requested.
	TargetSegmentDuration float64
	// Indicates whether to generate a single file (On-Demand profile) or
	// multiple files (Live profile). Default is false.
	GenerateDashIfIop bool
	// For DASH IOP. The default presented duration if it cannot be calculated
	// from the segments. This is in ISO 8601 duration format.
	DefaultPresentationDuration time.Duration
	// Indicates whether to use SegmentTimeline in SegmentTemplate. This flag
	// is always set to true for live profiles.
	UseSegmentTimeline bool
	// Indicates whether to allow approximate SegmentTimeline, where if
	// enabled, the SegmentTimeline does not need to be aligned to the
	// actual media segment boundaries.
	AllowApproximateSegmentTimeline bool
}

// NewMpdParams creates a MpdParams with default values.
func NewMpdParams() MpdParams {
	return MpdParams{
		GenerateStaticLiveMpd:            false,
		MinimumUpdatePeriod:              0,
		TimeShiftBufferDepth:             0,
		SuggestedPresentationDelay:       0,
		MinBufferTime:                    2.0,
		TargetSegmentDuration:            0,
		GenerateDashIfIop:                false,
		UseSegmentTimeline:               true,
		AllowApproximateSegmentTimeline:  false,
	}
}