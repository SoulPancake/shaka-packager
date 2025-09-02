package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/SoulPancake/shaka-packager/go-wrapper"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run example.go <input-file>")
		fmt.Println("Example: go run example.go input.mp4")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Fatalf("Input file does not exist: %s", inputFile)
	}

	// Get library version
	version := gopackager.GetLibraryVersion()
	fmt.Printf("Shaka Packager version: %s\n", version)

	// Create packager instance
	packager := gopackager.NewPackager()
	if packager == nil {
		log.Fatal("Failed to create packager instance")
	}
	defer packager.Close()

	// Create output directory
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Configure packaging parameters
	params := gopackager.PackagingParams{
		TempDir:         filepath.Join(outputDir, "temp"),
		OutputMediaInfo: true,
		SingleThreaded:  true, // For deterministic output in example
		Mp4OutputParams: gopackager.Mp4OutputParams{
			IncludePsshInStream: false,
		},
		ChunkingParams: gopackager.ChunkingParams{
			SegmentDurationInSeconds: 6.0, // 6 second segments
		},
	}

	// Define input/output streams
	streams := []gopackager.StreamDescriptor{
		{
			Input:          inputFile,
			StreamSelector: "video",
			Output:         filepath.Join(outputDir, "video.mp4"),
		},
		{
			Input:          inputFile,
			StreamSelector: "audio",
			Output:         filepath.Join(outputDir, "audio.mp4"),
		},
	}

	fmt.Printf("Initializing packager with input: %s\n", inputFile)
	fmt.Printf("Output directory: %s\n", outputDir)

	// Initialize the packager
	if err := packager.Initialize(params, streams); err != nil {
		log.Fatalf("Failed to initialize packager: %v", err)
	}

	fmt.Println("Running packaging operation...")

	// Run the packaging operation
	if err := packager.Run(); err != nil {
		log.Fatalf("Packaging failed: %v", err)
	}

	fmt.Println("Packaging completed successfully!")
	fmt.Printf("Output files:\n")
	fmt.Printf("  - Video: %s\n", filepath.Join(outputDir, "video.mp4"))
	fmt.Printf("  - Audio: %s\n", filepath.Join(outputDir, "audio.mp4"))
	
	if params.OutputMediaInfo {
		fmt.Printf("  - Media info files: %s/*.media_info\n", outputDir)
	}
}