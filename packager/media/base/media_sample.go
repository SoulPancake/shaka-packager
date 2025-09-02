// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package base

import (
	"fmt"
	"sync"
)

// DecryptConfig represents decryption configuration for encrypted media
type DecryptConfig struct {
	KeyId               []byte
	Iv                  []byte
	SubsampleEncryption bool
	ProtectionScheme    string
}

// MediaSample holds a media sample (frame/packet)
type MediaSample struct {
	mu                    sync.RWMutex
	data                  []byte
	sideData              []byte
	dts                   int64  // Decode timestamp 
	pts                   int64  // Presentation timestamp
	duration              int64  // Duration of the sample
	isKeyFrame            bool
	isEncrypted           bool
	encryptConfig         *DecryptConfig
	configId              int64
	isEOS                 bool   // End of stream marker
	isMetadata            bool   // Whether this is metadata only
}

// CopyFrom creates a MediaSample from input data
func CopyFrom(data []byte, isKeyFrame bool) *MediaSample {
	sample := &MediaSample{
		data:       make([]byte, len(data)),
		isKeyFrame: isKeyFrame,
	}
	copy(sample.data, data)
	return sample
}

// CopyFromWithSideData creates a MediaSample from input data with side data
func CopyFromWithSideData(data []byte, sideData []byte, isKeyFrame bool) *MediaSample {
	sample := &MediaSample{
		data:       make([]byte, len(data)),
		sideData:   make([]byte, len(sideData)),
		isKeyFrame: isKeyFrame,
	}
	copy(sample.data, data)
	copy(sample.sideData, sideData)
	return sample
}

// FromMetadata creates a MediaSample from metadata only
func FromMetadata(metadata []byte) *MediaSample {
	sample := &MediaSample{
		data:       make([]byte, len(metadata)),
		isKeyFrame: false,
		isMetadata: true,
	}
	copy(sample.data, metadata)
	return sample
}

// CreateEmptyMediaSample creates an empty media sample
func CreateEmptyMediaSample() *MediaSample {
	return &MediaSample{}
}

// CreateEOSBuffer creates an end-of-stream marker sample
func CreateEOSBuffer() *MediaSample {
	return &MediaSample{
		isEOS: true,
	}
}

// Clone creates a copy of this media sample
func (ms *MediaSample) Clone() *MediaSample {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	clone := &MediaSample{
		data:       make([]byte, len(ms.data)),
		sideData:   make([]byte, len(ms.sideData)),
		dts:        ms.dts,
		pts:        ms.pts,
		duration:   ms.duration,
		isKeyFrame: ms.isKeyFrame,
		isEncrypted: ms.isEncrypted,
		configId:   ms.configId,
		isEOS:      ms.isEOS,
		isMetadata: ms.isMetadata,
	}
	
	copy(clone.data, ms.data)
	copy(clone.sideData, ms.sideData)
	
	if ms.encryptConfig != nil {
		clone.encryptConfig = &DecryptConfig{
			KeyId:               make([]byte, len(ms.encryptConfig.KeyId)),
			Iv:                  make([]byte, len(ms.encryptConfig.Iv)),
			SubsampleEncryption: ms.encryptConfig.SubsampleEncryption,
			ProtectionScheme:    ms.encryptConfig.ProtectionScheme,
		}
		copy(clone.encryptConfig.KeyId, ms.encryptConfig.KeyId)
		copy(clone.encryptConfig.Iv, ms.encryptConfig.Iv)
	}
	
	return clone
}

// TransferData transfers data to this media sample without copying
func (ms *MediaSample) TransferData(data []byte) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data = data
}

// SetData copies data to this media sample
func (ms *MediaSample) SetData(data []byte) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data = make([]byte, len(data))
	copy(ms.data, data)
}

// ToString returns a human-readable string describing this sample
func (ms *MediaSample) ToString() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.isEOS {
		return "EOS"
	}
	
	return fmt.Sprintf("MediaSample{pts: %d, dts: %d, duration: %d, size: %d, keyframe: %v, encrypted: %v}",
		ms.pts, ms.dts, ms.duration, len(ms.data), ms.isKeyFrame, ms.isEncrypted)
}

// Getters and setters

// DTS returns the decode timestamp
func (ms *MediaSample) DTS() int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.dts
}

// SetDTS sets the decode timestamp
func (ms *MediaSample) SetDTS(dts int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.dts = dts
}

// PTS returns the presentation timestamp
func (ms *MediaSample) PTS() int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.pts
}

// SetPTS sets the presentation timestamp
func (ms *MediaSample) SetPTS(pts int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.pts = pts
}

// Duration returns the sample duration
func (ms *MediaSample) Duration() int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.duration
}

// SetDuration sets the sample duration
func (ms *MediaSample) SetDuration(duration int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.duration = duration
}

// IsKeyFrame returns whether this is a key frame
func (ms *MediaSample) IsKeyFrame() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.isKeyFrame
}

// SetIsKeyFrame sets the key frame flag
func (ms *MediaSample) SetIsKeyFrame(isKeyFrame bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.isKeyFrame = isKeyFrame
}

// IsEncrypted returns whether this sample is encrypted
func (ms *MediaSample) IsEncrypted() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.isEncrypted
}

// IsEOS returns whether this is an end-of-stream marker
func (ms *MediaSample) IsEOS() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.isEOS
}

// IsMetadata returns whether this sample contains only metadata
func (ms *MediaSample) IsMetadata() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.isMetadata
}

// Data returns the sample data
func (ms *MediaSample) Data() []byte {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.data
}

// SideData returns the side data
func (ms *MediaSample) SideData() []byte {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.sideData
}

// Size returns the size of the sample data
func (ms *MediaSample) Size() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.data)
}

// SideDataSize returns the size of the side data
func (ms *MediaSample) SideDataSize() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.sideData)
}

// ConfigId returns the config ID
func (ms *MediaSample) ConfigId() int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.configId
}

// SetConfigId sets the config ID
func (ms *MediaSample) SetConfigId(configId int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.configId = configId
}

// DecryptConfig returns the decryption configuration
func (ms *MediaSample) DecryptConfig() *DecryptConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.encryptConfig
}

// SetDecryptConfig sets the decryption configuration
func (ms *MediaSample) SetDecryptConfig(config *DecryptConfig) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.encryptConfig = config
	ms.isEncrypted = config != nil
}