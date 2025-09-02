// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style  
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

// Classes defines common class-like patterns in Go.
// This replaces the C++ macros for class definitions.

// NonCopyable provides a way to prevent copying of structs.
// Embed this struct in any struct that should not be copied.
type NonCopyable struct {
	// This field prevents copying by making the zero value invalid
	noCopy noCopy
}

// noCopy may be embedded into structs which must not be copied after the first use.
// This is a more idiomatic Go way to prevent copying.
type noCopy struct{}

// Lock is a no-op method to prevent copying detection by go vet.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// DisallowCopyAndAssign is a Go equivalent of the C++ DISALLOW_COPY_AND_ASSIGN macro.
// In Go, we achieve this through interface design and embedding NonCopyable.

// Example usage:
// type MyStruct struct {
//     NonCopyable
//     // other fields...
// }