// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package app

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

// FieldType represents the type of a stream descriptor field
type FieldType int

const (
	UnknownField FieldType = iota
	StreamSelectorField
	InputField
	OutputField
	SegmentTemplateField
	BandwidthField
	LanguageField
	CcIndexField
	OutputFormatField
	HlsNameField
	HlsGroupIdField
	HlsPlaylistNameField
	HlsIframePlaylistNameField
	TrickPlayFactorField
	SkipEncryptionField
	DrmStreamLabelField
	HlsCharacteristicsField
	DashAccessibilitiesField
	DashRolesField
	DashOnlyField
	HlsOnlyField
	DashLabelField
	ForcedSubtitleField
	InputFormatField
)

// FieldNameMapping maps field names to their types
type FieldNameMapping struct {
	FieldName string
	FieldType FieldType
}

// Field name to type mappings
var fieldNameTypeMappings = []FieldNameMapping{
	{"stream_selector", StreamSelectorField},
	{"stream", StreamSelectorField},
	{"input", InputField},
	{"in", InputField},
	{"output", OutputField},
	{"out", OutputField},
	{"init_segment", OutputField},
	{"segment_template", SegmentTemplateField},
	{"template", SegmentTemplateField},
	{"bandwidth", BandwidthField},
	{"bw", BandwidthField},
	{"bitrate", BandwidthField},
	{"language", LanguageField},
	{"lang", LanguageField},
	{"cc_index", CcIndexField},
	{"output_format", OutputFormatField},
	{"format", OutputFormatField},
	{"hls_name", HlsNameField},
	{"hls_group_id", HlsGroupIdField},
	{"playlist_name", HlsPlaylistNameField},
	{"iframe_playlist_name", HlsIframePlaylistNameField},
	{"trick_play_factor", TrickPlayFactorField},
	{"tpf", TrickPlayFactorField},
	{"skip_encryption", SkipEncryptionField},
	{"drm_stream_label", DrmStreamLabelField},
	{"drm_label", DrmStreamLabelField},
	{"hls_characteristics", HlsCharacteristicsField},
	{"characteristics", HlsCharacteristicsField},
	{"charcs", HlsCharacteristicsField},
	{"dash_accessibilities", DashAccessibilitiesField},
	{"dash_accessibility", DashAccessibilitiesField},
	{"accessibilities", DashAccessibilitiesField},
	{"accessibility", DashAccessibilitiesField},
	{"dash_roles", DashRolesField},
	{"dash_role", DashRolesField},
	{"roles", DashRolesField},
	{"role", DashRolesField},
	{"dash_only", DashOnlyField},
	{"hls_only", HlsOnlyField},
	{"dash_label", DashLabelField},
	{"forced_subtitle", ForcedSubtitleField},
	{"input_format", InputFormatField},
}

// GetFieldType returns the field type for a given field name
func GetFieldType(fieldName string) FieldType {
	for _, mapping := range fieldNameTypeMappings {
		if fieldName == mapping.FieldName {
			return mapping.FieldType
		}
	}
	return UnknownField
}

// ParseStreamDescriptor parses a descriptor string into a StreamDescriptor
// The descriptor string contains comma-separated name-value pairs describing the stream
func ParseStreamDescriptor(descriptorString string) (*packager.StreamDescriptor, error) {
	if descriptorString == "" {
		return nil, errors.New("empty descriptor string")
	}
	
	descriptor := &packager.StreamDescriptor{}
	
	// Split the descriptor string by commas
	pairs := strings.Split(descriptorString, ",")
	
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		// Split each pair by '='
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key-value pair: %s", pair)
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		if err := setDescriptorField(descriptor, key, value); err != nil {
			return nil, fmt.Errorf("error setting field %s: %v", key, err)
		}
	}
	
	// Validate required fields
	if err := validateDescriptor(descriptor); err != nil {
		return nil, fmt.Errorf("descriptor validation failed: %v", err)
	}
	
	return descriptor, nil
}

// setDescriptorField sets a field in the stream descriptor based on field type
func setDescriptorField(descriptor *packager.StreamDescriptor, key, value string) error {
	fieldType := GetFieldType(key)
	
	switch fieldType {
	case StreamSelectorField:
		descriptor.Stream = value
		
	case InputField:
		descriptor.Input = value
		
	case OutputField:
		descriptor.Output = value
		
	case SegmentTemplateField:
		descriptor.SegmentTemplate = value
		
	case BandwidthField:
		bandwidth, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid bandwidth: %s", value)
		}
		descriptor.Bandwidth = uint32(bandwidth)
		
	case LanguageField:
		descriptor.Language = value
		
	case CcIndexField:
		ccIndex, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid cc_index: %s", value)
		}
		descriptor.CcIndex = ccIndex
		
	case OutputFormatField:
		descriptor.OutputFormat = value
		
	case HlsNameField:
		descriptor.HlsName = value
		
	case HlsGroupIdField:
		descriptor.HlsGroupId = value
		
	case HlsPlaylistNameField:
		descriptor.HlsPlaylistName = value
		
	case HlsIframePlaylistNameField:
		descriptor.HlsIframePlaylistName = value
		
	case TrickPlayFactorField:
		factor, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid trick_play_factor: %s", value)
		}
		descriptor.TrickPlayFactor = factor
		
	case SkipEncryptionField:
		skipEncryption, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid skip_encryption: %s", value)
		}
		descriptor.SkipEncryption = skipEncryption
		
	case DrmStreamLabelField:
		descriptor.DrmStreamLabel = value
		
	case HlsCharacteristicsField:
		// Parse comma-separated characteristics
		characteristics := strings.Split(value, ";")
		for i, char := range characteristics {
			characteristics[i] = strings.TrimSpace(char)
		}
		descriptor.HlsCharacteristics = characteristics
		
	case DashAccessibilitiesField:
		// Parse comma-separated accessibility values
		accessibilities := strings.Split(value, ";")
		for i, acc := range accessibilities {
			accessibilities[i] = strings.TrimSpace(acc)
		}
		descriptor.DashAccessibilities = accessibilities
		
	case DashRolesField:
		// Parse comma-separated roles
		roles := strings.Split(value, ";")
		for i, role := range roles {
			roles[i] = strings.TrimSpace(role)
		}
		descriptor.DashRoles = roles
		
	case DashOnlyField:
		dashOnly, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid dash_only: %s", value)
		}
		descriptor.DashOnly = dashOnly
		
	case HlsOnlyField:
		hlsOnly, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid hls_only: %s", value)
		}
		descriptor.HlsOnly = hlsOnly
		
	case DashLabelField:
		descriptor.DashLabel = value
		
	case ForcedSubtitleField:
		forcedSubtitle, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid forced_subtitle: %s", value)
		}
		descriptor.ForcedSubtitle = forcedSubtitle
		
	case InputFormatField:
		descriptor.InputFormat = value
		
	case UnknownField:
		log.Printf("Warning: unknown field %s with value %s", key, value)
		
	default:
		return fmt.Errorf("unsupported field type for %s", key)
	}
	
	return nil
}

// validateDescriptor validates that the descriptor has required fields
func validateDescriptor(descriptor *packager.StreamDescriptor) error {
	if descriptor.Input == "" {
		return errors.New("input field is required")
	}
	
	if descriptor.Output == "" && descriptor.SegmentTemplate == "" {
		return errors.New("either output or segment_template is required")
	}
	
	// Additional validation can be added here
	
	return nil
}

// ParseStreamDescriptors parses multiple stream descriptor strings
func ParseStreamDescriptors(descriptorStrings []string) ([]*packager.StreamDescriptor, error) {
	var descriptors []*packager.StreamDescriptor
	
	for i, descriptorString := range descriptorStrings {
		descriptor, err := ParseStreamDescriptor(descriptorString)
		if err != nil {
			return nil, fmt.Errorf("error parsing stream descriptor %d: %v", i+1, err)
		}
		descriptors = append(descriptors, descriptor)
	}
	
	return descriptors, nil
}