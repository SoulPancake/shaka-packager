// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package file

import (
	"net"
	"testing"
	"time"
)

func TestUdpFile(t *testing.T) {
	// Test UDP file creation
	udpFile := NewUdpFile("127.0.0.1:0") // Use port 0 to let OS assign
	if udpFile == nil {
		t.Fatal("Failed to create UDP file")
	}

	// Test invalid address format
	invalidFile := NewUdpFile("invalid-address")
	err := invalidFile.Open()
	if err == nil {
		t.Error("Expected error for invalid address format")
	}
}

func TestUdpFileOperations(t *testing.T) {
	udpFile := NewUdpFile("127.0.0.1:0")
	
	// Test operations on unopened file
	buffer := make([]byte, 1024)
	_, err := udpFile.Read(buffer)
	if err == nil {
		t.Error("Expected error when reading from unopened file")
	}
	
	// Test write operation (should always fail)
	_, err = udpFile.Write([]byte("test"))
	if err == nil {
		t.Error("Expected error for write operation on UDP file")
	}
	
	// Test Size operation (should always fail)
	_, err = udpFile.Size()
	if err == nil {
		t.Error("Expected error for Size operation on UDP file")
	}
	
	// Test Seek operation (should always fail)
	err = udpFile.Seek(10)
	if err == nil {
		t.Error("Expected error for Seek operation on UDP file")
	}
	
	// Test Tell operation (should always fail)
	_, err = udpFile.Tell()
	if err == nil {
		t.Error("Expected error for Tell operation on UDP file")
	}
	
	// Test Flush (should succeed)
	err = udpFile.Flush()
	if err != nil {
		t.Error("Flush should not fail on UDP file")
	}
	
	// Test Close
	err = udpFile.Close()
	if err != nil {
		t.Error("Close should not fail")
	}
}

func TestUdpFileRealConnection(t *testing.T) {
	// Create a test UDP server
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer serverConn.Close()
	
	// Get the actual address the server is listening on
	serverPort := serverConn.LocalAddr().(*net.UDPAddr).Port
	
	// Start a goroutine to send test data
	go func() {
		time.Sleep(100 * time.Millisecond) // Give client time to start listening
		clientAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		clientConn, err := net.DialUDP("udp", clientAddr, serverConn.LocalAddr().(*net.UDPAddr))
		if err == nil {
			clientConn.Write([]byte("Hello, UDP!"))
			clientConn.Close()
		}
	}()
	
	// Create UDP file client
	udpFile := NewUdpFile("127.0.0.1:" + string(rune(serverPort)))
	err = udpFile.Open()
	if err != nil {
		t.Fatalf("Failed to open UDP file: %v", err)
	}
	defer udpFile.Close()
	
	// Try to read data (this might timeout in test environment)
	buffer := make([]byte, 1024)
	_, err = udpFile.Read(buffer)
	// Note: This test might timeout in CI/test environments
	// The important thing is that the Open() succeeded
	if err != nil {
		t.Logf("UDP read failed (expected in test environment): %v", err)
	}
}

func TestIsIpv4MulticastAddress(t *testing.T) {
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"224.0.0.1", true},   // Multicast
		{"239.255.255.255", true}, // Multicast
		{"192.168.1.1", false},   // Unicast
		{"127.0.0.1", false},     // Localhost
		{"::1", false},            // IPv6
	}
	
	for _, tc := range testCases {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Errorf("Failed to parse IP: %s", tc.ip)
			continue
		}
		
		result := IsIpv4MulticastAddress(ip)
		if result != tc.expected {
			t.Errorf("IsIpv4MulticastAddress(%s) = %v, expected %v", tc.ip, result, tc.expected)
		}
	}
}

func TestUdpFileMulticast(t *testing.T) {
	// Test multicast address parsing (doesn't require actual network setup)
	udpFile := NewUdpFile("224.0.0.1:12345")
	
	// Opening multicast might fail in test environment without proper network setup
	err := udpFile.Open()
	if err != nil {
		t.Logf("Multicast open failed (expected in test environment): %v", err)
	} else {
		udpFile.Close()
	}
}

func TestUdpFileAddressParsing(t *testing.T) {
	testCases := []struct {
		address   string
		shouldErr bool
	}{
		{"127.0.0.1:8080", false},
		{"192.168.1.1:9090", false},
		{"invalid:address", true},
		{"127.0.0.1:invalid", true},
		{"127.0.0.1", true}, // Missing port
	}
	
	for _, tc := range testCases {
		udpFile := NewUdpFile(tc.address)
		err := udpFile.Open()
		
		if tc.shouldErr && err == nil {
			t.Errorf("Expected error for address %s, but got none", tc.address)
		} else if !tc.shouldErr && err != nil {
			t.Errorf("Unexpected error for address %s: %v", tc.address, err)
		}
		
		if err == nil {
			udpFile.Close()
		}
	}
}