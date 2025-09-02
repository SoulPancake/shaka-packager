// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package packager provides the core implementation of the shaka packager.
package packager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"../include/packager"
	"./status"
)

// PackagerInternal contains the internal state of the packager.
type PackagerInternal struct {
	mu                   sync.RWMutex
	ctx                  context.Context
	cancel               context.CancelFunc
	packagingParams      packager.PackagingParams
	streamDescriptors    []packager.StreamDescriptor
	isInitialized        bool
	isRunning            bool
	outputFiles          []string
	tempFiles            []string
	startTime            time.Time
	endTime              time.Time
}

// Packager is the main packager implementation.
type Packager struct {
	internal *PackagerInternal
}

// New creates a new Packager instance.
func New() *Packager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Packager{
		internal: &PackagerInternal{
			ctx:    ctx,
			cancel: cancel,
		},
	}
}

// Initialize initializes the packaging pipeline.
func (p *Packager) Initialize(packagingParams packager.PackagingParams, streamDescriptors []packager.StreamDescriptor) packager.Status {
	p.internal.mu.Lock()
	defer p.internal.mu.Unlock()

	if p.internal.isInitialized {
		return packager.NewStatus(packager.INTERNAL_ERROR, "Packager already initialized")
	}

	if p.internal.isRunning {
		return packager.NewStatus(packager.INTERNAL_ERROR, "Packager is running")
	}

	// Validate parameters
	if err := p.validateParameters(packagingParams, streamDescriptors); err != nil {
		return err
	}

	// Store parameters
	p.internal.packagingParams = packagingParams
	p.internal.streamDescriptors = streamDescriptors
	p.internal.isInitialized = true

	return packager.StatusOK
}

// Run runs the pipeline to completion.
func (p *Packager) Run() packager.Status {
	p.internal.mu.Lock()
	
	if !p.internal.isInitialized {
		p.internal.mu.Unlock()
		return packager.NewStatus(packager.INTERNAL_ERROR, "Packager not initialized")
	}

	if p.internal.isRunning {
		p.internal.mu.Unlock()
		return packager.NewStatus(packager.INTERNAL_ERROR, "Packager already running")
	}

	p.internal.isRunning = true
	p.internal.startTime = time.Now()
	p.internal.mu.Unlock()

	defer func() {
		p.internal.mu.Lock()
		p.internal.isRunning = false
		p.internal.endTime = time.Now()
		p.internal.mu.Unlock()
	}()

	// Run the packaging pipeline
	return p.runPackaging()
}

// Cancel cancels the packaging operation.
func (p *Packager) Cancel() {
	p.internal.mu.Lock()
	defer p.internal.mu.Unlock()
	
	if p.internal.cancel != nil {
		p.internal.cancel()
	}
}

// Close cleans up resources.
func (p *Packager) Close() error {
	p.Cancel()
	
	// Clean up temporary files
	p.internal.mu.Lock()
	tempFiles := make([]string, len(p.internal.tempFiles))
	copy(tempFiles, p.internal.tempFiles)
	p.internal.mu.Unlock()
	
	for _, tempFile := range tempFiles {
		packager.GlobalFileUtils.Delete(tempFile)
	}
	
	return nil
}

// validateParameters validates the packaging parameters and stream descriptors.
func (p *Packager) validateParameters(params packager.PackagingParams, descriptors []packager.StreamDescriptor) packager.Status {
	if len(descriptors) == 0 {
		return packager.NewStatus(packager.INVALID_ARGUMENT, "No stream descriptors provided")
	}

	for i, desc := range descriptors {
		if desc.Input == "" {
			return packager.NewStatus(packager.INVALID_ARGUMENT, fmt.Sprintf("Stream descriptor %d missing input", i))
		}
		
		if desc.StreamSelector == "" {
			return packager.NewStatus(packager.INVALID_ARGUMENT, fmt.Sprintf("Stream descriptor %d missing stream selector", i))
		}

		// Validate output format
		if desc.Output != "" && desc.OutputFormat == "" {
			// Try to detect format from output file extension
			ext := strings.ToLower(filepath.Ext(desc.Output))
			switch ext {
			case ".mp4":
				// Will be set during processing
			case ".m3u8":
				// Will be set during processing
			case ".mpd":
				// Will be set during processing
			default:
				return packager.NewStatus(packager.INVALID_ARGUMENT, 
					fmt.Sprintf("Stream descriptor %d: cannot detect output format from file extension %s", i, ext))
			}
		}
	}

	// Validate encryption parameters
	if params.EncryptionParams.KeyProvider != packager.KeyProviderNone {
		if err := p.validateEncryptionParams(params.EncryptionParams); err != nil {
			return err
		}
	}

	return packager.StatusOK
}

// validateEncryptionParams validates encryption parameters.
func (p *Packager) validateEncryptionParams(params packager.EncryptionParams) packager.Status {
	switch params.KeyProvider {
	case packager.KeyProviderRawKey:
		if len(params.RawKey.KeyMap) == 0 {
			return packager.NewStatus(packager.INVALID_ARGUMENT, "Raw key encryption requires key map")
		}
	case packager.KeyProviderWidevine:
		if params.Widevine.KeyServerURL == "" {
			return packager.NewStatus(packager.INVALID_ARGUMENT, "Widevine encryption requires key server URL")
		}
	case packager.KeyProviderPlayReady:
		if params.PlayReady.KeyServerURL == "" {
			return packager.NewStatus(packager.INVALID_ARGUMENT, "PlayReady encryption requires key server URL")
		}
		if params.PlayReady.ProgramIdentifier == "" {
			return packager.NewStatus(packager.INVALID_ARGUMENT, "PlayReady encryption requires program identifier")
		}
	}
	
	return packager.StatusOK
}

// runPackaging runs the actual packaging pipeline.
func (p *Packager) runPackaging() packager.Status {
	// Create processing pipeline for each stream
	var wg sync.WaitGroup
	statusChan := make(chan packager.Status, len(p.internal.streamDescriptors))

	for i, desc := range p.internal.streamDescriptors {
		wg.Add(1)
		go func(index int, descriptor packager.StreamDescriptor) {
			defer wg.Done()
			
			select {
			case <-p.internal.ctx.Done():
				statusChan <- packager.NewStatus(packager.CANCELLED, "Operation cancelled")
				return
			default:
			}
			
			// Process individual stream
			status := p.processStream(index, descriptor)
			statusChan <- status
		}(i, desc)
	}

	// Wait for all streams to complete
	wg.Wait()
	close(statusChan)

	// Check for errors
	var firstError packager.Status
	for status := range statusChan {
		if !status.OK() && firstError.OK() {
			firstError = status
		}
	}

	if !firstError.OK() {
		return firstError
	}

	// Generate manifests if needed
	return p.generateManifests()
}

// processStream processes a single stream descriptor.
func (p *Packager) processStream(index int, descriptor packager.StreamDescriptor) packager.Status {
	// TODO: Implement stream processing pipeline
	// This would include:
	// 1. Opening input file
	// 2. Demuxing
	// 3. Stream selection
	// 4. Encryption (if enabled)
	// 5. Chunking/Segmentation
	// 6. Muxing
	// 7. Writing output

	// For now, return success
	return packager.StatusOK
}

// generateManifests generates DASH MPD and HLS playlists.
func (p *Packager) generateManifests() packager.Status {
	// Generate DASH MPD if requested
	if p.internal.packagingParams.MpdParams.MasterPlaylistOutput != "" {
		if err := p.generateDashMpd(); err != nil {
			return err
		}
	}

	// Generate HLS master playlist if requested  
	if p.internal.packagingParams.HlsParams.MasterPlaylistOutput != "" {
		if err := p.generateHlsPlaylist(); err != nil {
			return err
		}
	}

	return packager.StatusOK
}

// generateDashMpd generates DASH MPD manifest.
func (p *Packager) generateDashMpd() packager.Status {
	// TODO: Implement DASH MPD generation
	return packager.StatusOK
}

// generateHlsPlaylist generates HLS master playlist.
func (p *Packager) generateHlsPlaylist() packager.Status {
	// TODO: Implement HLS playlist generation
	return packager.StatusOK
}

// GetLibraryVersion returns the version of the library.
func GetLibraryVersion() string {
	return packager.LibraryVersion
}

// DefaultStreamLabelFunction provides the default stream label function.
func DefaultStreamLabelFunction(maxSDPixels, maxHDPixels, maxUHD1Pixels int, streamAttributes packager.EncryptedStreamAttributes) string {
	return packager.DefaultStreamLabelFunction(maxSDPixels, maxHDPixels, maxUHD1Pixels, streamAttributes)
}