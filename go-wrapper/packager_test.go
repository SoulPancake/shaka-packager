package gopackager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// PackagerTestSuite provides a test suite with common setup/teardown
type PackagerTestSuite struct {
	suite.Suite
	tempDir     string
	testDataDir string
}

// SetupSuite runs once before all tests in the suite
func (suite *PackagerTestSuite) SetupSuite() {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "packager_test_*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Create test data directory
	suite.testDataDir = filepath.Join(suite.tempDir, "test_data")
	err = os.MkdirAll(suite.testDataDir, 0755)
	suite.Require().NoError(err)

	// Create a mock input file for testing
	mockInputFile := filepath.Join(suite.testDataDir, "test_input.mp4")
	err = ioutil.WriteFile(mockInputFile, []byte("mock mp4 content"), 0644)
	suite.Require().NoError(err)
}

// TearDownSuite runs once after all tests in the suite
func (suite *PackagerTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// SetupTest runs before each individual test
func (suite *PackagerTestSuite) SetupTest() {
	// Create a fresh subdirectory for each test
	testName := suite.T().Name()
	testDir := filepath.Join(suite.tempDir, strings.ReplaceAll(testName, "/", "_"))
	err := os.MkdirAll(testDir, 0755)
	suite.Require().NoError(err)
}

// TestNewPackager tests packager creation and destruction
func (suite *PackagerTestSuite) TestNewPackager() {
	// Test successful creation
	packager := NewPackager()
	suite.NotNil(packager, "NewPackager should return a valid packager instance")
	suite.NotNil(packager.ptr, "Packager should have a valid internal pointer")

	// Test that Close() works without error
	suite.NotPanics(func() {
		packager.Close()
	}, "Close should not panic")

	// Test that calling Close() twice is safe
	suite.NotPanics(func() {
		packager.Close()
	}, "Calling Close twice should be safe")

	// Test that ptr is nil after Close
	suite.Nil(packager.ptr, "Pointer should be nil after Close()")
}

// TestPackagerMultipleInstances tests creating multiple packager instances
func (suite *PackagerTestSuite) TestPackagerMultipleInstances() {
	const numInstances = 10
	packagers := make([]*Packager, numInstances)

	// Create multiple instances
	for i := 0; i < numInstances; i++ {
		packagers[i] = NewPackager()
		suite.NotNil(packagers[i], "Each packager instance should be valid")
	}

	// Clean up all instances
	for i := 0; i < numInstances; i++ {
		suite.NotPanics(func() {
			packagers[i].Close()
		}, "Each instance should close without panic")
	}
}

// TestGetLibraryVersion tests getting the library version
func (suite *PackagerTestSuite) TestGetLibraryVersion() {
	version := GetLibraryVersion()
	suite.NotEmpty(version, "Library version should not be empty")
	suite.NotEqual("unknown", version, "Library version should be determinable")

	// Version should contain some expected patterns
	// Note: This is a basic check - actual version format may vary
	suite.True(len(version) > 0, "Version string should have content")
}

// TestPackagingParamsValidation tests parameter validation
func (suite *PackagerTestSuite) TestPackagingParamsValidation() {
	packager := NewPackager()
	defer packager.Close()

	// Test with empty streams - should fail
	params := PackagingParams{
		TempDir:         suite.tempDir,
		OutputMediaInfo: true,
		SingleThreaded:  true,
	}

	err := packager.Initialize(params, []StreamDescriptor{})
	suite.Error(err, "Initialize with empty streams should fail")
	suite.Contains(err.Error(), "no streams specified", "Error should mention empty streams")
}

// TestStreamDescriptorValidation tests stream descriptor validation
func (suite *PackagerTestSuite) TestStreamDescriptorValidation() {
	packager := NewPackager()
	defer packager.Close()

	params := PackagingParams{
		TempDir:         suite.tempDir,
		OutputMediaInfo: true,
		SingleThreaded:  true,
	}

	// Test with invalid stream (empty input)
	streams := []StreamDescriptor{
		{
			Input:          "", // Empty input should cause issues
			StreamSelector: "video",
			Output:         filepath.Join(suite.tempDir, "output.mp4"),
		},
	}

	_ = packager.Initialize(params, streams)
	// Note: This may or may not fail depending on C++ validation
	// We just ensure the function doesn't panic
	suite.NotPanics(func() {
		packager.Initialize(params, streams)
	}, "Initialize should not panic with invalid stream")
}

// TestValidStreamDescriptor tests initialization with valid parameters
func (suite *PackagerTestSuite) TestValidStreamDescriptor() {
	packager := NewPackager()
	defer packager.Close()

	params := PackagingParams{
		TempDir:         suite.tempDir,
		OutputMediaInfo: true,
		SingleThreaded:  true,
	}

	streams := []StreamDescriptor{
		{
			Input:          filepath.Join(suite.testDataDir, "test_input.mp4"),
			StreamSelector: "video",
			Output:         filepath.Join(suite.tempDir, "output_video.mp4"),
		},
		{
			Input:          filepath.Join(suite.testDataDir, "test_input.mp4"),
			StreamSelector: "audio",
			Output:         filepath.Join(suite.tempDir, "output_audio.mp4"),
		},
	}

	// This should not panic, though it may fail due to invalid input file format
	err := packager.Initialize(params, streams)
	suite.NotPanics(func() {
		packager.Initialize(params, streams)
	}, "Initialize should not panic with valid parameters")

	// If initialization succeeded, test that we can call other methods
	if err == nil {
		// Test Cancel (should not panic)
		suite.NotPanics(func() {
			packager.Cancel()
		}, "Cancel should not panic")

		// Test Run (may fail, but should not panic)
		suite.NotPanics(func() {
			packager.Run()
		}, "Run should not panic")
	}
}

// TestPackagerAfterClose tests that methods fail appropriately after Close
func (suite *PackagerTestSuite) TestPackagerAfterClose() {
	packager := NewPackager()
	packager.Close()

	params := PackagingParams{
		TempDir: suite.tempDir,
	}
	streams := []StreamDescriptor{
		{
			Input:          "input.mp4",
			StreamSelector: "video",
			Output:         "output.mp4",
		},
	}

	// Test that Initialize fails after Close
	err := packager.Initialize(params, streams)
	suite.Error(err, "Initialize should fail after Close")
	suite.Contains(err.Error(), "not initialized", "Error should mention initialization")

	// Test that Run fails after Close
	err = packager.Run()
	suite.Error(err, "Run should fail after Close")
	suite.Contains(err.Error(), "not initialized", "Error should mention initialization")

	// Test that Cancel is safe after Close
	suite.NotPanics(func() {
		packager.Cancel()
	}, "Cancel should be safe after Close")
}

// TestPackagerConcurrentAccess tests concurrent access to packager methods
func (suite *PackagerTestSuite) TestPackagerConcurrentAccess() {
	packager := NewPackager()
	defer packager.Close()

	var wg sync.WaitGroup
	const numGoroutines = 5

	// Test concurrent Cancel calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			packager.Cancel()
		}()
	}

	wg.Wait()

	// Test concurrent GetLibraryVersion calls
	versions := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			versions[index] = GetLibraryVersion()
		}(i)
	}

	wg.Wait()

	// All versions should be the same
	baseVersion := versions[0]
	for i := 1; i < numGoroutines; i++ {
		suite.Equal(baseVersion, versions[i], "All concurrent GetLibraryVersion calls should return the same result")
	}
}

// TestStatusCodeMapping tests that status codes are properly mapped to errors
func (suite *PackagerTestSuite) TestStatusCodeMapping() {
	// Test the statusToError function with different codes
	testCases := []struct {
		code     int
		expected string
	}{
		{StatusOK, ""},
		{StatusUnknown, "unknown error"},
		{StatusCancelled, "operation cancelled"},
		{StatusInvalidArgument, "invalid argument"},
		{StatusNotFound, "not found"},
		{StatusAlreadyExists, "already exists"},
		{StatusResourceExhausted, "resource exhausted"},
		{StatusFailedPrecondition, "failed precondition"},
		{StatusAborted, "aborted"},
		{StatusOutOfRange, "out of range"},
		{StatusUnimplemented, "unimplemented"},
		{StatusInternal, "internal error"},
		{StatusUnavailable, "unavailable"},
		{StatusDataLoss, "data loss"},
		{StatusUnauthenticated, "unauthenticated"},
		{999, "unknown status code: 999"},
	}

	for _, tc := range testCases {
		err := statusToError(tc.code)
		if tc.expected == "" {
			suite.NoError(err, "StatusOK should not produce an error")
		} else {
			suite.Error(err, "Non-OK status should produce an error")
			suite.Contains(err.Error(), tc.expected, "Error message should contain expected text")
		}
	}
}

// TestBoolToIntConversion tests the boolToInt helper function
func (suite *PackagerTestSuite) TestBoolToIntConversion() {
	// Test the helper function through public interface
	// We can't directly test boolToInt as it's internal, but we can verify behavior
	
	packager := NewPackager()
	defer packager.Close()
	
	params1 := PackagingParams{SingleThreaded: true}
	params2 := PackagingParams{SingleThreaded: false}
	
	// These should not panic when using boolean values
	suite.NotPanics(func() {
		packager.Initialize(params1, []StreamDescriptor{{Input: "test", StreamSelector: "video"}})
	}, "Should handle true boolean")
	
	suite.NotPanics(func() {
		packager.Initialize(params2, []StreamDescriptor{{Input: "test", StreamSelector: "video"}})
	}, "Should handle false boolean")
}

// TestPackagingParamsFields tests that all parameter fields can be set
func (suite *PackagerTestSuite) TestPackagingParamsFields() {
	params := PackagingParams{
		TempDir:         "/tmp/test",
		OutputMediaInfo: true,
		SingleThreaded:  true,
		Mp4OutputParams: Mp4OutputParams{
			IncludePsshInStream:           true,
			GenerateDashIfIopCompliantMpd: true,
		},
		ChunkingParams: ChunkingParams{
			SegmentDurationInSeconds:    10.0,
			SubsegmentDurationInSeconds: 2.0,
		},
		EncryptionParams: EncryptionParams{
			KeyId:              "0123456789abcdef0123456789abcdef",
			Key:                "fedcba9876543210fedcba9876543210",
			ClearLeadInSeconds: 5.0,
		},
	}

	// Verify all fields are accessible
	suite.Equal("/tmp/test", params.TempDir)
	suite.True(params.OutputMediaInfo)
	suite.True(params.SingleThreaded)
	suite.True(params.Mp4OutputParams.IncludePsshInStream)
	suite.True(params.Mp4OutputParams.GenerateDashIfIopCompliantMpd)
	suite.Equal(10.0, params.ChunkingParams.SegmentDurationInSeconds)
	suite.Equal(2.0, params.ChunkingParams.SubsegmentDurationInSeconds)
	suite.Equal("0123456789abcdef0123456789abcdef", params.EncryptionParams.KeyId)
	suite.Equal("fedcba9876543210fedcba9876543210", params.EncryptionParams.Key)
	suite.Equal(5.0, params.EncryptionParams.ClearLeadInSeconds)
}

// TestStreamDescriptorFields tests that all stream descriptor fields work
func (suite *PackagerTestSuite) TestStreamDescriptorFields() {
	index := uint32(1)
	stream := StreamDescriptor{
		Input:           "/path/to/input.mp4",
		StreamSelector:  "video",
		Output:          "/path/to/output.mp4",
		SegmentTemplate: "segment_$Number$.m4s",
		Index:           &index,
	}

	// Verify all fields are accessible
	suite.Equal("/path/to/input.mp4", stream.Input)
	suite.Equal("video", stream.StreamSelector)
	suite.Equal("/path/to/output.mp4", stream.Output)
	suite.Equal("segment_$Number$.m4s", stream.SegmentTemplate)
	suite.NotNil(stream.Index)
	suite.Equal(uint32(1), *stream.Index)
}

// TestNilPackagerHandling tests handling of nil packager pointers
func (suite *PackagerTestSuite) TestNilPackagerHandling() {
	// Create a packager with nil pointer (simulating creation failure)
	packager := &Packager{ptr: nil}

	params := PackagingParams{TempDir: suite.tempDir}
	streams := []StreamDescriptor{
		{Input: "test.mp4", StreamSelector: "video", Output: "output.mp4"},
	}

	// Test that methods handle nil pointer gracefully
	err := packager.Initialize(params, streams)
	suite.Error(err, "Initialize should fail with nil pointer")

	err = packager.Run()
	suite.Error(err, "Run should fail with nil pointer")

	suite.NotPanics(func() {
		packager.Cancel()
	}, "Cancel should not panic with nil pointer")

	suite.NotPanics(func() {
		packager.Close()
	}, "Close should not panic with nil pointer")
}

// TestLargeStreamList tests handling of many streams
func (suite *PackagerTestSuite) TestLargeStreamList() {
	packager := NewPackager()
	defer packager.Close()

	params := PackagingParams{
		TempDir:        suite.tempDir,
		SingleThreaded: true,
	}

	// Create a large number of streams
	const numStreams = 100
	streams := make([]StreamDescriptor, numStreams)
	for i := 0; i < numStreams; i++ {
		streams[i] = StreamDescriptor{
			Input:          filepath.Join(suite.testDataDir, "test_input.mp4"),
			StreamSelector: fmt.Sprintf("stream_%d", i),
			Output:         filepath.Join(suite.tempDir, fmt.Sprintf("output_%d.mp4", i)),
		}
	}

	// This should not panic, though it may fail due to invalid streams
	suite.NotPanics(func() {
		packager.Initialize(params, streams)
	}, "Initialize should handle large stream lists without panicking")
}

// TestMemoryManagement tests that memory is properly managed
func (suite *PackagerTestSuite) TestMemoryManagement() {
	// Create and destroy many packagers to test for memory leaks
	const iterations = 1000

	for i := 0; i < iterations; i++ {
		packager := NewPackager()
		suite.NotNil(packager, "Each packager should be created successfully")
		packager.Close()
	}

	// Test with initialization
	for i := 0; i < 10; i++ {
		packager := NewPackager()
		suite.NotNil(packager)

		params := PackagingParams{TempDir: suite.tempDir}
		streams := []StreamDescriptor{
			{Input: "test.mp4", StreamSelector: "video", Output: "output.mp4"},
		}

		// Initialize may fail, but should not leak memory
		packager.Initialize(params, streams)
		packager.Close()
	}
}

// TestTimeoutBehavior tests behavior under timeout conditions
func (suite *PackagerTestSuite) TestTimeoutBehavior() {
	packager := NewPackager()
	defer packager.Close()

	params := PackagingParams{
		TempDir:        suite.tempDir,
		SingleThreaded: true,
	}

	streams := []StreamDescriptor{
		{
			Input:          filepath.Join(suite.testDataDir, "test_input.mp4"),
			StreamSelector: "video",
			Output:         filepath.Join(suite.tempDir, "output.mp4"),
		},
	}

	// Initialize (may fail, but should complete quickly)
	done := make(chan bool, 1)
	go func() {
		packager.Initialize(params, streams)
		done <- true
	}()

	select {
	case <-done:
		// Completed normally
	case <-time.After(5 * time.Second):
		suite.Fail("Initialize took longer than expected")
	}

	// Test Cancel from another goroutine
	go func() {
		time.Sleep(100 * time.Millisecond)
		packager.Cancel()
	}()

	// Run should be interruptible
	suite.NotPanics(func() {
		packager.Run()
	}, "Run should be interruptible and not panic")
}

// Run all tests in the suite
func TestPackagerSuite(t *testing.T) {
	suite.Run(t, new(PackagerTestSuite))
}

// Benchmark tests

// BenchmarkPackagerCreation benchmarks packager creation and destruction
func BenchmarkPackagerCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		packager := NewPackager()
		if packager != nil {
			packager.Close()
		}
	}
}

// BenchmarkGetLibraryVersion benchmarks version retrieval
func BenchmarkGetLibraryVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetLibraryVersion()
	}
}

// BenchmarkStatusToError benchmarks error conversion
func BenchmarkStatusToError(b *testing.B) {
	statuses := []int{StatusOK, StatusUnknown, StatusInvalidArgument, StatusInternal}
	for i := 0; i < b.N; i++ {
		_ = statusToError(statuses[i%len(statuses)])
	}
}

// Example tests demonstrating usage

// ExamplePackager demonstrates basic packager usage
func ExamplePackager() {
	// Create a new packager instance
	packager := NewPackager()
	defer packager.Close()

	// Configure packaging parameters
	params := PackagingParams{
		TempDir:         "/tmp/packaging",
		OutputMediaInfo: true,
		SingleThreaded:  true, // For deterministic output
	}

	// Define input/output streams
	streams := []StreamDescriptor{
		{
			Input:          "input_video.mp4",
			StreamSelector: "video",
			Output:         "output_video.mp4",
		},
		{
			Input:          "input_video.mp4",
			StreamSelector: "audio",
			Output:         "output_audio.mp4",
		},
	}

	// Initialize the packager
	if err := packager.Initialize(params, streams); err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		return
	}

	// Run the packaging operation
	if err := packager.Run(); err != nil {
		fmt.Printf("Packaging failed: %v\n", err)
		return
	}

	fmt.Println("Packaging completed successfully")
}

// ExampleGetLibraryVersion demonstrates version retrieval
func ExampleGetLibraryVersion() {
	version := GetLibraryVersion()
	fmt.Printf("Shaka Packager version: %s\n", version)
}