// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package utils provides utility functions for the shaka packager.
package utils

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Clock interface provides time-related functionality.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

// RealClock provides real system time.
type RealClock struct{}

// Now returns the current time.
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// Since returns the time elapsed since t.
func (c *RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// TestClock provides fake time for testing.
type TestClock struct {
	currentTime time.Time
}

// NewTestClock creates a new test clock starting at the given time.
func NewTestClock(startTime time.Time) *TestClock {
	return &TestClock{currentTime: startTime}
}

// Now returns the current fake time.
func (c *TestClock) Now() time.Time {
	return c.currentTime
}

// Since returns the fake time elapsed since t.
func (c *TestClock) Since(t time.Time) time.Duration {
	return c.currentTime.Sub(t)
}

// Advance advances the fake time by the given duration.
func (c *TestClock) Advance(d time.Duration) {
	c.currentTime = c.currentTime.Add(d)
}

// SetTime sets the fake time to the given time.
func (c *TestClock) SetTime(t time.Time) {
	c.currentTime = t
}

// Default clock instance
var DefaultClock Clock = &RealClock{}