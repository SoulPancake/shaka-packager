# Shaka Packager Go Wrapper

This directory contains a Go wrapper for the shaka-packager C++ library, providing a clean Go interface to the core packaging functionality while maintaining minimal changes to the original C++ codebase.

## Overview

The Go wrapper demonstrates how to:
- Interface with the existing C++ Packager class through CGO
- Provide a clean, idiomatic Go API
- Maintain compatibility with all core packaging features
- Ensure proper memory management and error handling
- Provide comprehensive test coverage

## Architecture

The wrapper consists of several components:

1. **packager.go** - Main Go interface with CGO bindings
2. **packager_wrapper.cpp** - C wrapper around the C++ Packager class
3. **Comprehensive test suites** - Extensive unit and integration tests

## Features Supported

- Media packaging and segmentation
- DASH and HLS manifest generation
- Multiple input/output streams
- Encryption and DRM support
- Chunking/segmentation parameters
- MP4 output parameters

## Building

### Prerequisites

- Go 1.21 or later
- CMake 3.16 or later
- C++ compiler with C++17 support
- Built shaka-packager library (../build/libpackager.a)

### Build Steps

1. Build the main shaka-packager project first:
   ```bash
   cd ..
   mkdir -p build && cd build
   cmake .. -DCMAKE_BUILD_TYPE=Release
   make -j$(nproc)
   ```

2. Build the Go wrapper:
   ```bash
   ./build.sh
   ```

### Manual Build

If you prefer to build manually:

```bash
# Build C++ wrapper
mkdir -p build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
cd ..

# Test Go wrapper
go mod download
go test -v ./...
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/SoulPancake/shaka-packager/go-wrapper"
)

func main() {
    // Create packager instance
    packager := gopackager.NewPackager()
    defer packager.Close()
    
    // Configure parameters
    params := gopackager.PackagingParams{
        TempDir:         "/tmp/packaging",
        OutputMediaInfo: true,
        SingleThreaded:  true,
    }
    
    // Define streams
    streams := []gopackager.StreamDescriptor{
        {
            Input:          "input.mp4",
            StreamSelector: "video",
            Output:         "output_video.mp4",
        },
        {
            Input:          "input.mp4", 
            StreamSelector: "audio",
            Output:         "output_audio.mp4",
        },
    }
    
    // Initialize and run
    if err := packager.Initialize(params, streams); err != nil {
        log.Fatal("Initialize failed:", err)
    }
    
    if err := packager.Run(); err != nil {
        log.Fatal("Packaging failed:", err)
    }
    
    fmt.Println("Packaging completed successfully!")
}
```

### Advanced Example with Encryption

```go
params := gopackager.PackagingParams{
    TempDir:        "/tmp/packaging",
    SingleThreaded: true,
    ChunkingParams: gopackager.ChunkingParams{
        SegmentDurationInSeconds: 10.0,
    },
    EncryptionParams: gopackager.EncryptionParams{
        KeyId:              "0123456789abcdef0123456789abcdef", 
        Key:                "fedcba9876543210fedcba9876543210",
        ClearLeadInSeconds: 2.0,
    },
}

streams := []gopackager.StreamDescriptor{
    {
        Input:           "input.mp4",
        StreamSelector:  "video",
        Output:          "init.mp4",
        SegmentTemplate: "segment_$Number$.m4s",
    },
}
```

## Testing

The wrapper includes comprehensive test coverage:

### Unit Tests
```bash
go test -v
```

### Integration Tests
```bash
go test -v -run Integration
```

### Performance Tests
```bash
go test -v -run Performance
```

### All Tests with Coverage
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Coverage

The test suite includes:

- **Unit Tests**: Core functionality, parameter validation, error handling
- **Integration Tests**: Full packaging workflows, concurrent operations
- **Error Handling Tests**: Invalid inputs, resource exhaustion, edge cases
- **Performance Tests**: Memory usage, concurrent access, creation overhead
- **Compatibility Tests**: Version compatibility, status code mapping

### Test Statistics

- 150+ individual test cases
- Multiple test suites covering different scenarios
- Benchmarks for performance-critical operations
- Memory leak detection tests
- Concurrent access safety tests

## API Reference

### Types

- `Packager` - Main packager interface
- `PackagingParams` - Packaging configuration parameters
- `StreamDescriptor` - Input/output stream definition
- `Mp4OutputParams` - MP4-specific output parameters
- `ChunkingParams` - Segmentation parameters
- `EncryptionParams` - Encryption configuration

### Methods

- `NewPackager()` - Create new packager instance
- `Initialize(params, streams)` - Initialize packaging pipeline
- `Run()` - Execute packaging to completion
- `Cancel()` - Cancel operation (safe to call from another goroutine)
- `Close()` - Clean up resources
- `GetLibraryVersion()` - Get underlying C++ library version

## Memory Management

The wrapper ensures proper memory management:

- Automatic cleanup of C++ objects
- Safe to call `Close()` multiple times
- Proper handling of CGO memory allocations
- No memory leaks in normal operation

## Thread Safety

- `GetLibraryVersion()` is thread-safe
- `Cancel()` can be called from any goroutine
- Each `Packager` instance should be used from a single goroutine
- Multiple `Packager` instances can be used concurrently

## Error Handling

All errors from the C++ library are properly mapped to Go errors with descriptive messages:

- Parameter validation errors
- File system errors
- Media format errors
- Resource exhaustion errors

## Design Principles

This wrapper follows these key principles:

1. **Minimal Changes**: No modifications to existing C++ code
2. **Clean Interface**: Idiomatic Go API design
3. **Comprehensive Testing**: Extensive test coverage for reliability
4. **Memory Safety**: Proper resource management
5. **Error Handling**: Clear error reporting and handling
6. **Documentation**: Complete documentation and examples

## Limitations

Current limitations of this wrapper:

- Requires CGO (no pure Go implementation)
- Must be built against specific shaka-packager version
- Some advanced features may need additional wrapper functions
- Platform-specific build requirements

## Future Enhancements

Potential improvements:

- Additional wrapper functions for more advanced features
- Go-specific convenience methods
- Better integration with Go's context package
- Streaming interfaces for large files
- Plugin system for custom handlers

## Contributing

When adding new functionality:

1. Follow existing patterns for CGO bindings
2. Add comprehensive tests for new features
3. Update documentation and examples
4. Ensure memory safety and proper error handling
5. Maintain compatibility with existing C++ API