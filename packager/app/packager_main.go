// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SoulPancake/shaka-packager/include/packager"
)

// Command line flags
var (
	dumpStreamInfo = flag.Bool("dump_stream_info", false, "Dump demuxed stream info")
	licenses       = flag.Bool("licenses", false, "Dump licenses")
	quiet          = flag.Bool("quiet", false, "When enabled, LOG(INFO) output is suppressed")
	useFakeClock   = flag.Bool("use_fake_clock_for_muxer", false, 
		"Set to true to use a fake clock for muxer. With this flag set, "+
		"creation time and modification time in outputs are set to 0. "+
		"Should only be used for testing.")
	testPackagerVersion = flag.String("test_packager_version", "", 
		"Packager version for testing. Should be used for testing only.")
	singleThreaded = flag.Bool("single_threaded", false, 
		"If enabled, only use one thread when generating content.")
)

// PackagerMain is the main entry point for the packager application
func PackagerMain(args []string) int {
	// Parse command line flags
	flag.CommandLine.Parse(args[1:])
	
	// Handle special flags
	if *licenses {
		printLicenses()
		return 0
	}
	
	if *quiet {
		log.SetOutput(os.Stderr) // Suppress INFO logs but keep errors
	}
	
	// Get remaining arguments (stream descriptors)
	streamDescriptors := flag.Args()
	
	if len(streamDescriptors) == 0 {
		printUsage()
		return 1
	}
	
	if *dumpStreamInfo {
		return dumpStreamInfo(streamDescriptors)
	}
	
	return runPackager(streamDescriptors)
}

// printUsage prints the usage information
func printUsage() {
	fmt.Printf(`Shaka Packager version %s
Usage: packager [flags] <stream_descriptor> ...

Stream Descriptors:
  Stream descriptors are comma-separated key-value pairs that describe the input and output.
  
  Example:
    packager \
      input=input.mp4,stream=audio,output=audio.mp4 \
      input=input.mp4,stream=video,output=video.mp4

Common flags:
`, getPackagerVersion())
	
	flag.PrintDefaults()
}

// printLicenses prints license information
func printLicenses() {
	fmt.Println("Shaka Packager uses the following third-party libraries:")
	fmt.Println("- Go standard library (BSD-style license)")
	fmt.Println("- Various Go packages with permissive licenses")
	fmt.Println("")
	fmt.Println("For complete license information, see the LICENSE file.")
}

// getPackagerVersion returns the packager version
func getPackagerVersion() string {
	if *testPackagerVersion != "" {
		return *testPackagerVersion
	}
	return "1.0.0" // Default version
}

// dumpStreamInfo dumps information about the input streams
func dumpStreamInfo(streamDescriptors []string) int {
	fmt.Println("Stream Information:")
	
	for i, descriptor := range streamDescriptors {
		fmt.Printf("Stream %d: %s\n", i+1, descriptor)
		
		// Parse the descriptor to extract input file
		pairs := parseStreamDescriptor(descriptor)
		if input, ok := pairs["input"]; ok {
			fmt.Printf("  Input: %s\n", input)
			// In a full implementation, we would analyze the input file here
			// and print codec information, duration, etc.
		}
	}
	
	return 0
}

// parseStreamDescriptor parses a stream descriptor string into key-value pairs
func parseStreamDescriptor(descriptor string) map[string]string {
	pairs := make(map[string]string)
	
	parts := strings.Split(descriptor, ",")
	for _, part := range parts {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			pairs[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	
	return pairs
}

// runPackager runs the main packaging operation
func runPackager(streamDescriptors []string) int {
	fmt.Printf("Starting Shaka Packager %s\n", getPackagerVersion())
	
	// Create packager instance
	pkg, err := packager.NewPackager()
	if err != nil {
		log.Printf("Failed to create packager: %v", err)
		return 1
	}
	defer pkg.Close()
	
	// Set up packager parameters
	params := &packager.PackagingParams{
		SingleThreaded: *singleThreaded,
		UseFakeClock:   *useFakeClock,
	}
	
	// Initialize packager
	err = pkg.Initialize(params, streamDescriptors)
	if err != nil {
		log.Printf("Failed to initialize packager: %v", err)
		return 1
	}
	
	// Run packaging
	err = pkg.Run()
	if err != nil {
		log.Printf("Packaging failed: %v", err)
		return 1
	}
	
	fmt.Println("Packaging completed successfully")
	return 0
}

// Main function for when used as a standalone executable
func Main() {
	os.Exit(PackagerMain(os.Args))
}