// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package origin

import (
	"github.com/SoulPancake/shaka-packager/packager/media/base"
	"github.com/SoulPancake/shaka-packager/packager/status"
)

// OriginHandler is a handler that sits at the head of a pipeline
// They take input from an alternative source (like a file or network connection)
type OriginHandler interface {
	base.MediaHandler
	
	// Run processes all data and sends messages downstream
	// This is the main method of the handler. Since origin handlers do not take
	// input via Process, Run will take input from an alternative source.
	// This call is expected to be blocking. To exit a call to Run, Cancel should be used.
	Run() *status.Status
	
	// Cancel is a non-blocking call requesting that the handler exit the current
	// call to Run. The handler should stop processing data as soon as convenient.
	Cancel()
}

// BaseOriginHandler provides a base implementation of OriginHandler
type BaseOriginHandler struct {
	*base.BaseMediaHandler
	cancelRequested chan bool
}

// NewBaseOriginHandler creates a new BaseOriginHandler
func NewBaseOriginHandler() *BaseOriginHandler {
	return &BaseOriginHandler{
		BaseMediaHandler: base.NewBaseMediaHandler(),
		cancelRequested:  make(chan bool, 1),
	}
}

// Process overrides the base Process method to prevent input via Process
// Origin handlers should not receive input through Process
func (boh *BaseOriginHandler) Process(streamData *base.StreamData) *status.Status {
	return status.NewStatus(status.InvalidArgument, 
		"origin handlers should not receive input through Process")
}

// Run is the main processing method that should be implemented by concrete handlers
func (boh *BaseOriginHandler) Run() *status.Status {
	// Base implementation - should be overridden by concrete handlers
	return status.NewStatus(status.Unimplemented, "Run method must be implemented by concrete handlers")
}

// Cancel requests that the handler stop processing
func (boh *BaseOriginHandler) Cancel() {
	// Signal cancellation
	select {
	case boh.cancelRequested <- true:
	default:
		// Already cancelled
	}
}

// IsCancelled returns true if cancellation has been requested
func (boh *BaseOriginHandler) IsCancelled() bool {
	select {
	case <-boh.cancelRequested:
		return true
	default:
		return false
	}
}

// SendStreamData sends stream data to the next handler in the pipeline
func (boh *BaseOriginHandler) SendStreamData(streamData *base.StreamData) *status.Status {
	return boh.Dispatch(streamData)
}