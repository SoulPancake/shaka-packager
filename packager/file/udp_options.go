// Copyright 2016 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// UdpOptions represents options parsed from UDP url string of the form: udp://ip:port[?options]
type UdpOptions struct {
	// IP Address
	address string
	port    uint16
	// Allow or disallow reusing UDP sockets
	reuse bool
	// Address of the interface over which to receive UDP multicast streams
	interfaceAddress string
	// Timeout in microseconds. 0 to indicate unlimited timeout
	timeoutUs uint32
	// Source specific multicast source address
	sourceAddress string
	// Whether this is source specific multicast
	isSourceSpecificMulticast bool
	// Maximum receive buffer size in bytes
	bufferSize int
}

// NewUdpOptions creates a new UdpOptions with default values
func NewUdpOptions() *UdpOptions {
	return &UdpOptions{
		address:          "0.0.0.0",
		port:             0,
		reuse:            false,
		interfaceAddress: "0.0.0.0",
		timeoutUs:        0,
		sourceAddress:    "0.0.0.0",
		isSourceSpecificMulticast: false,
		bufferSize:       0,
	}
}

// ParseFromString parses UDP options from URL string
// udpUrl should be of the form udp://ip:port[?options]
// Returns a UdpOptions object on success, error otherwise
func ParseFromString(udpUrl string) (*UdpOptions, error) {
	if !strings.HasPrefix(udpUrl, "udp://") {
		return nil, fmt.Errorf("invalid UDP URL scheme: %s", udpUrl)
	}
	
	parsedUrl, err := url.Parse(udpUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UDP URL: %v", err)
	}
	
	options := NewUdpOptions()
	
	// Extract host and port
	options.address = parsedUrl.Hostname()
	if parsedUrl.Port() != "" {
		port, err := strconv.ParseUint(parsedUrl.Port(), 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %v", err)
		}
		options.port = uint16(port)
	}
	
	// Parse query parameters
	queryParams := parsedUrl.Query()
	
	// Parse reuse option
	if reuse := queryParams.Get("reuse"); reuse != "" {
		options.reuse = strings.ToLower(reuse) == "true" || reuse == "1"
	}
	
	// Parse interface address
	if iface := queryParams.Get("interface"); iface != "" {
		options.interfaceAddress = iface
	}
	
	// Parse timeout
	if timeout := queryParams.Get("timeout"); timeout != "" {
		timeoutVal, err := strconv.ParseUint(timeout, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout value: %v", err)
		}
		options.timeoutUs = uint32(timeoutVal)
	}
	
	// Parse source address for source-specific multicast
	if source := queryParams.Get("source"); source != "" {
		options.sourceAddress = source
		options.isSourceSpecificMulticast = true
	}
	
	// Parse buffer size
	if bufSize := queryParams.Get("buffer_size"); bufSize != "" {
		size, err := strconv.Atoi(bufSize)
		if err != nil {
			return nil, fmt.Errorf("invalid buffer size: %v", err)
		}
		options.bufferSize = size
	}
	
	return options, nil
}

// Address returns the IP address
func (u *UdpOptions) Address() string {
	return u.address
}

// Port returns the port number  
func (u *UdpOptions) Port() uint16 {
	return u.port
}

// Reuse returns whether socket reuse is enabled
func (u *UdpOptions) Reuse() bool {
	return u.reuse
}

// InterfaceAddress returns the interface address for multicast
func (u *UdpOptions) InterfaceAddress() string {
	return u.interfaceAddress
}

// TimeoutUs returns the timeout in microseconds
func (u *UdpOptions) TimeoutUs() uint32 {
	return u.timeoutUs
}

// SourceAddress returns the source address for source-specific multicast
func (u *UdpOptions) SourceAddress() string {
	return u.sourceAddress
}

// IsSourceSpecificMulticast returns whether this is source-specific multicast
func (u *UdpOptions) IsSourceSpecificMulticast() bool {
	return u.isSourceSpecificMulticast
}

// BufferSize returns the buffer size
func (u *UdpOptions) BufferSize() int {
	return u.bufferSize
}