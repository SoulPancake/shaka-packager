// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// BufferCallbackFunc defines the callback function for buffer operations.
type BufferCallbackFunc func(name string, data []byte) error

// BufferCallbackParams contains parameters for buffer callbacks.
type BufferCallbackParams struct {
	// Callback function to be called when buffer operations occur.
	BufferCallback BufferCallbackFunc
	// User data to be passed to the callback function.
	UserData interface{}
}

// NewBufferCallbackParams creates a BufferCallbackParams with default values.
func NewBufferCallbackParams() BufferCallbackParams {
	return BufferCallbackParams{}
}