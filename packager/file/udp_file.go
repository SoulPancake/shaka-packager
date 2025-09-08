// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UdpFile implements File interface for receiving UDP unicast and multicast streams
type UdpFile struct {
	mu       sync.RWMutex
	fileName string
	conn     net.PacketConn
	closed   bool
	addr     *net.UDPAddr
}

// NewUdpFile creates a new UDP file
// address should be of the form "<ip_address>:<port>"
func NewUdpFile(address string) *UdpFile {
	return &UdpFile{
		fileName: address,
	}
}

// Close closes the UDP connection
func (uf *UdpFile) Close() error {
	uf.mu.Lock()
	defer uf.mu.Unlock()
	
	if uf.closed {
		return nil
	}
	
	if uf.conn != nil {
		err := uf.conn.Close()
		uf.conn = nil
		if err != nil {
			return fmt.Errorf("failed to close UDP connection: %v", err)
		}
	}
	
	uf.closed = true
	return nil
}

// Read reads data from the UDP connection
func (uf *UdpFile) Read(buffer []byte) (int64, error) {
	uf.mu.RLock()
	defer uf.mu.RUnlock()
	
	if uf.closed {
		return 0, errors.New("UDP file is closed")
	}
	
	if uf.conn == nil {
		return 0, errors.New("UDP connection not established")
	}
	
	if len(buffer) < 65535 {
		log.Printf("Buffer may be too small to read entire UDP datagram")
	}
	
	// Set read timeout to avoid blocking indefinitely
	uf.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	
	n, _, err := uf.conn.ReadFrom(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return 0, errors.New("UDP read timeout")
		}
		return 0, fmt.Errorf("UDP read error: %v", err)
	}
	
	return int64(n), nil
}

// Write is not supported for UDP file (read-only)
func (uf *UdpFile) Write(buffer []byte) (int64, error) {
	return 0, errors.New("UdpFile is unwritable")
}

// CloseForWriting does nothing for UDP file
func (uf *UdpFile) CloseForWriting() {
	// Do nothing - UDP file is read-only
}

// Size is not supported for UDP streams
func (uf *UdpFile) Size() (int64, error) {
	return 0, errors.New("UdpFile does not support Size")
}

// Flush does nothing for UDP file
func (uf *UdpFile) Flush() error {
	return nil // No buffering to flush
}

// Seek is not supported for UDP streams
func (uf *UdpFile) Seek(position uint64) error {
	return errors.New("UdpFile does not support Seek")
}

// Tell is not supported for UDP streams  
func (uf *UdpFile) Tell() (uint64, error) {
	return 0, errors.New("UdpFile does not support Tell")
}

// Open establishes the UDP connection
func (uf *UdpFile) Open() error {
	uf.mu.Lock()
	defer uf.mu.Unlock()
	
	if uf.closed {
		return errors.New("UDP file is closed")
	}
	
	// Parse the address
	host, portStr, err := net.SplitHostPort(uf.fileName)
	if err != nil {
		return fmt.Errorf("invalid UDP address format: %v", err)
	}
	
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port number: %v", err)
	}
	
	// Create UDP address
	uf.addr = &net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}
	
	if uf.addr.IP == nil {
		return fmt.Errorf("invalid IP address: %s", host)
	}
	
	// Check if this is a multicast address
	if uf.addr.IP.IsMulticast() {
		// Join multicast group
		conn, err := net.ListenMulticastUDP("udp", nil, uf.addr)
		if err != nil {
			return fmt.Errorf("failed to join multicast group: %v", err)
		}
		uf.conn = conn
	} else {
		// Regular UDP listen
		conn, err := net.ListenUDP("udp", uf.addr)
		if err != nil {
			return fmt.Errorf("failed to listen on UDP address: %v", err)
		}
		uf.conn = conn
	}
	
	log.Printf("UDP connection established on %s", uf.fileName)
	return nil
}

// IsIpv4MulticastAddress checks if the IP address is IPv4 multicast
func IsIpv4MulticastAddress(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] >= 224 && ip4[0] <= 239
	}
	return false
}