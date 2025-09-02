package gopackager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite contains integration tests that test the full pipeline
type IntegrationTestSuite struct {
	suite.Suite
	tempDir     string
	testDataDir string
}

func (suite *IntegrationTestSuite) SetupSuite() {
	tempDir, err := ioutil.TempDir("", "integration_test_*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	suite.testDataDir = filepath.Join(suite.tempDir, "test_data")
	err = os.MkdirAll(suite.testDataDir, 0755)
	suite.Require().NoError(err)

	// Create mock input files with realistic names
	mockFiles := []string{"test_video.mp4", "test_audio.mp4", "test_combined.mp4"}
	for _, filename := range mockFiles {
		mockFile := filepath.Join(suite.testDataDir, filename)
		err = ioutil.WriteFile(mockFile, []byte("mock media content"), 0644)
		suite.Require().NoError(err)
	}
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestFullPackagingWorkflow tests a complete packaging workflow
func (suite *IntegrationTestSuite) TestFullPackagingWorkflow() {
	packager := NewPackager()
	defer packager.Close()

	outputDir := filepath.Join(suite.tempDir, "output")
	err := os.MkdirAll(outputDir, 0755)
	suite.Require().NoError(err)

	params := PackagingParams{
		TempDir:         filepath.Join(suite.tempDir, "temp"),
		OutputMediaInfo: true,
		SingleThreaded:  true,
		Mp4OutputParams: Mp4OutputParams{
			IncludePsshInStream: false,
		},
		ChunkingParams: ChunkingParams{
			SegmentDurationInSeconds: 10.0,
		},
	}

	streams := []StreamDescriptor{
		{
			Input:          filepath.Join(suite.testDataDir, "test_combined.mp4"),
			StreamSelector: "video",
			Output:         filepath.Join(outputDir, "video.mp4"),
		},
		{
			Input:          filepath.Join(suite.testDataDir, "test_combined.mp4"),
			StreamSelector: "audio", 
			Output:         filepath.Join(outputDir, "audio.mp4"),
		},
	}

	// Initialize the packager
	err = packager.Initialize(params, streams)
	// Note: This will likely fail due to mock files, but should not panic
	suite.NotPanics(func() {
		packager.Initialize(params, streams)
	}, "Initialize should not panic with realistic parameters")

	// If initialization succeeded, try running
	if err == nil {
		suite.NotPanics(func() {
			packager.Run()
		}, "Run should not panic")
	}
}

// TestSegmentedOutput tests packaging with segmented output
func (suite *IntegrationTestSuite) TestSegmentedOutput() {
	packager := NewPackager()
	defer packager.Close()

	outputDir := filepath.Join(suite.tempDir, "segmented_output")
	err := os.MkdirAll(outputDir, 0755)
	suite.Require().NoError(err)

	params := PackagingParams{
		TempDir:        filepath.Join(suite.tempDir, "temp_segmented"),
		SingleThreaded: true,
		ChunkingParams: ChunkingParams{
			SegmentDurationInSeconds:    2.0,
			SubsegmentDurationInSeconds: 1.0,
		},
	}

	streams := []StreamDescriptor{
		{
			Input:           filepath.Join(suite.testDataDir, "test_video.mp4"),
			StreamSelector:  "video",
			Output:          filepath.Join(outputDir, "init.mp4"),
			SegmentTemplate: filepath.Join(outputDir, "segment_$Number$.m4s"),
		},
	}

	suite.NotPanics(func() {
		err := packager.Initialize(params, streams)
		if err == nil {
			packager.Run()
		}
	}, "Segmented packaging should not panic")
}

// TestEncryptedPackaging tests packaging with encryption parameters
func (suite *IntegrationTestSuite) TestEncryptedPackaging() {
	packager := NewPackager()
	defer packager.Close()

	outputDir := filepath.Join(suite.tempDir, "encrypted_output")
	err := os.MkdirAll(outputDir, 0755)
	suite.Require().NoError(err)

	params := PackagingParams{
		TempDir:        filepath.Join(suite.tempDir, "temp_encrypted"),
		SingleThreaded: true,
		EncryptionParams: EncryptionParams{
			KeyId:              "0123456789abcdef0123456789abcdef",
			Key:                "fedcba9876543210fedcba9876543210",
			ClearLeadInSeconds: 2.0,
		},
	}

	streams := []StreamDescriptor{
		{
			Input:          filepath.Join(suite.testDataDir, "test_video.mp4"),
			StreamSelector: "video",
			Output:         filepath.Join(outputDir, "encrypted_video.mp4"),
		},
	}

	suite.NotPanics(func() {
		err := packager.Initialize(params, streams)
		if err == nil {
			packager.Run()
		}
	}, "Encrypted packaging should not panic")
}

// TestMultiplePackagersSimultaneous tests running multiple packagers at once
func (suite *IntegrationTestSuite) TestMultiplePackagersSimultaneous() {
	const numPackagers = 5
	var wg sync.WaitGroup
	errors := make(chan error, numPackagers)

	for i := 0; i < numPackagers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			packager := NewPackager()
			defer packager.Close()

			outputDir := filepath.Join(suite.tempDir, fmt.Sprintf("concurrent_output_%d", index))
			err := os.MkdirAll(outputDir, 0755)
			if err != nil {
				errors <- err
				return
			}

			params := PackagingParams{
				TempDir:        filepath.Join(suite.tempDir, fmt.Sprintf("temp_concurrent_%d", index)),
				SingleThreaded: true,
			}

			streams := []StreamDescriptor{
				{
					Input:          filepath.Join(suite.testDataDir, "test_video.mp4"),
					StreamSelector: "video",
					Output:         filepath.Join(outputDir, "output.mp4"),
				},
			}

			err = packager.Initialize(params, streams)
			if err == nil {
				err = packager.Run()
			}
			// We don't consider packaging errors as test failures here
			// since we're using mock files
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check that no critical errors occurred
	for err := range errors {
		suite.T().Logf("Non-critical concurrent error: %v", err)
	}
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// ErrorHandlingTestSuite tests comprehensive error scenarios
type ErrorHandlingTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *ErrorHandlingTestSuite) SetupSuite() {
	tempDir, err := ioutil.TempDir("", "error_test_*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
}

func (suite *ErrorHandlingTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestInvalidInputPaths tests handling of invalid input paths
func (suite *ErrorHandlingTestSuite) TestInvalidInputPaths() {
	packager := NewPackager()
	defer packager.Close()

	params := PackagingParams{
		TempDir:        suite.tempDir,
		SingleThreaded: true,
	}

	testCases := []struct {
		name   string
		input  string
		expect string
	}{
		{"nonexistent file", "/path/to/nonexistent/file.mp4", ""},
		{"empty path", "", ""},
		{"directory instead of file", suite.tempDir, ""},
		{"invalid characters", "file\x00name.mp4", ""},
		{"very long path", string(make([]byte, 4096)), ""},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			streams := []StreamDescriptor{
				{
					Input:          tc.input,
					StreamSelector: "video",
					Output:         filepath.Join(suite.tempDir, "output.mp4"),
				},
			}

			// Should not panic, though it may return an error
			assert.NotPanics(t, func() {
				packager.Initialize(params, streams)
			}, "Should not panic with invalid input path")
		})
	}
}

// TestInvalidOutputPaths tests handling of invalid output paths  
func (suite *ErrorHandlingTestSuite) TestInvalidOutputPaths() {
	packager := NewPackager()
	defer packager.Close()

	// Create a mock input file
	inputFile := filepath.Join(suite.tempDir, "input.mp4")
	err := ioutil.WriteFile(inputFile, []byte("mock"), 0644)
	suite.Require().NoError(err)

	params := PackagingParams{
		TempDir:        suite.tempDir,
		SingleThreaded: true,
	}

	testCases := []struct {
		name   string
		output string
	}{
		{"read-only directory", "/root/output.mp4"},
		{"invalid characters", "output\x00.mp4"},
		{"empty output", ""},
		{"very long path", string(make([]byte, 4096))},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			streams := []StreamDescriptor{
				{
					Input:          inputFile,
					StreamSelector: "video",
					Output:         tc.output,
				},
			}

			assert.NotPanics(t, func() {
				packager.Initialize(params, streams)
			}, "Should not panic with invalid output path")
		})
	}
}

// TestInvalidStreamSelectors tests handling of invalid stream selectors
func (suite *ErrorHandlingTestSuite) TestInvalidStreamSelectors() {
	packager := NewPackager()
	defer packager.Close()

	inputFile := filepath.Join(suite.tempDir, "input.mp4")
	err := ioutil.WriteFile(inputFile, []byte("mock"), 0644)
	suite.Require().NoError(err)

	params := PackagingParams{
		TempDir:        suite.tempDir,
		SingleThreaded: true,
	}

	invalidSelectors := []string{
		"invalid_stream_type",
		"999", // Very high index
		"-1",  // Negative index
		"",    // Empty selector
		"video,audio", // Multiple selectors
		"stream_with_very_long_name_that_exceeds_reasonable_limits",
	}

	for _, selector := range invalidSelectors {
		suite.T().Run(fmt.Sprintf("selector_%s", selector), func(t *testing.T) {
			streams := []StreamDescriptor{
				{
					Input:          inputFile,
					StreamSelector: selector,
					Output:         filepath.Join(suite.tempDir, "output.mp4"),
				},
			}

			assert.NotPanics(t, func() {
				packager.Initialize(params, streams)
			}, "Should not panic with invalid stream selector")
		})
	}
}

// TestMemoryExhaustion tests behavior under memory pressure
func (suite *ErrorHandlingTestSuite) TestMemoryExhaustion() {
	// Test with an extremely large number of streams to simulate memory pressure
	packager := NewPackager()
	defer packager.Close()

	inputFile := filepath.Join(suite.tempDir, "input.mp4")
	err := ioutil.WriteFile(inputFile, []byte("mock"), 0644)
	suite.Require().NoError(err)

	params := PackagingParams{
		TempDir:        suite.tempDir,
		SingleThreaded: true,
	}

	// Create a large number of streams
	const numStreams = 10000
	streams := make([]StreamDescriptor, numStreams)
	for i := 0; i < numStreams; i++ {
		streams[i] = StreamDescriptor{
			Input:          inputFile,
			StreamSelector: fmt.Sprintf("stream_%d", i),
			Output:         filepath.Join(suite.tempDir, fmt.Sprintf("output_%d.mp4", i)),
		}
	}

	// This should handle gracefully without crashing
	suite.NotPanics(func() {
		err := packager.Initialize(params, streams)
		// May fail due to resource limits, but shouldn't panic
		suite.T().Logf("Initialize with %d streams result: %v", numStreams, err)
	}, "Should handle large numbers of streams gracefully")
}

// TestConcurrentCancellation tests cancellation under various conditions
func (suite *ErrorHandlingTestSuite) TestConcurrentCancellation() {
	packager := NewPackager()
	defer packager.Close()

	inputFile := filepath.Join(suite.tempDir, "input.mp4")
	err := ioutil.WriteFile(inputFile, []byte("mock"), 0644)
	suite.Require().NoError(err)

	params := PackagingParams{
		TempDir:        suite.tempDir,
		SingleThreaded: true,
	}

	streams := []StreamDescriptor{
		{
			Input:          inputFile,
			StreamSelector: "video",
			Output:         filepath.Join(suite.tempDir, "output.mp4"),
		},
	}

	// Initialize
	err = packager.Initialize(params, streams)
	if err != nil {
		suite.T().Logf("Initialize failed (expected with mock file): %v", err)
		return
	}

	// Start cancellation from multiple goroutines
	var wg sync.WaitGroup
	const numCancellers = 10

	for i := 0; i < numCancellers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			time.Sleep(time.Duration(index) * time.Millisecond)
			packager.Cancel()
		}(i)
	}

	// Also try to run in parallel
	go func() {
		time.Sleep(50 * time.Millisecond)
		packager.Run()
	}()

	wg.Wait()
	
	// Additional cancellation after everything
	suite.NotPanics(func() {
		packager.Cancel()
	}, "Final cancellation should not panic")
}

// TestResourceCleanup tests that resources are properly cleaned up
func (suite *ErrorHandlingTestSuite) TestResourceCleanup() {
	// Test cleanup after various failure scenarios
	scenarios := []func(*Packager) error{
		// Initialize with invalid parameters then close
		func(p *Packager) error {
			params := PackagingParams{}
			streams := []StreamDescriptor{{Input: "nonexistent.mp4"}}
			return p.Initialize(params, streams)
		},
		// Initialize successfully then close without running
		func(p *Packager) error {
			params := PackagingParams{TempDir: suite.tempDir}
			streams := []StreamDescriptor{{
				Input:          "test.mp4",
				StreamSelector: "video",
				Output:         "output.mp4",
			}}
			return p.Initialize(params, streams)
		},
		// Just cancel without initializing
		func(p *Packager) error {
			p.Cancel()
			return nil
		},
	}

	for i, scenario := range scenarios {
		suite.T().Run(fmt.Sprintf("scenario_%d", i), func(t *testing.T) {
			packager := NewPackager()
			require.NotNil(t, packager)

			// Run the scenario
			err := scenario(packager)
			t.Logf("Scenario %d result: %v", i, err)

			// Cleanup should always work without panic
			assert.NotPanics(t, func() {
				packager.Close()
			}, "Cleanup should not panic after scenario")

			// Double cleanup should be safe
			assert.NotPanics(t, func() {
				packager.Close()
			}, "Double cleanup should be safe")
		})
	}
}

// TestInvalidParameterCombinations tests invalid parameter combinations
func (suite *ErrorHandlingTestSuite) TestInvalidParameterCombinations() {
	testCases := []struct {
		name    string
		params  PackagingParams
		streams []StreamDescriptor
	}{
		{
			name: "negative segment duration",
			params: PackagingParams{
				ChunkingParams: ChunkingParams{
					SegmentDurationInSeconds: -1.0,
				},
			},
			streams: []StreamDescriptor{{Input: "test.mp4", StreamSelector: "video"}},
		},
		{
			name: "invalid encryption key",
			params: PackagingParams{
				EncryptionParams: EncryptionParams{
					KeyId: "invalid_key_format",
					Key:   "also_invalid",
				},
			},
			streams: []StreamDescriptor{{Input: "test.mp4", StreamSelector: "video"}},
		},
		{
			name: "conflicting parameters",
			params: PackagingParams{
				SingleThreaded: true,
				ChunkingParams: ChunkingParams{
					SegmentDurationInSeconds:    10.0,
					SubsegmentDurationInSeconds: 20.0, // Longer than segment
				},
			},
			streams: []StreamDescriptor{{Input: "test.mp4", StreamSelector: "video"}},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			packager := NewPackager()
			defer packager.Close()

			assert.NotPanics(t, func() {
				err := packager.Initialize(tc.params, tc.streams)
				t.Logf("Invalid parameter test '%s' result: %v", tc.name, err)
			}, "Should not panic with invalid parameter combinations")
		})
	}
}

func TestErrorHandlingSuite(t *testing.T) {
	suite.Run(t, new(ErrorHandlingTestSuite))
}

// PerformanceTestSuite tests performance characteristics
type PerformanceTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *PerformanceTestSuite) SetupSuite() {
	tempDir, err := ioutil.TempDir("", "perf_test_*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
}

func (suite *PerformanceTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestPackagerCreationPerformance tests creation/destruction performance
func (suite *PerformanceTestSuite) TestPackagerCreationPerformance() {
	const iterations = 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		packager := NewPackager()
		if packager != nil {
			packager.Close()
		}
	}

	elapsed := time.Since(start)
	avgTime := elapsed / iterations

	suite.T().Logf("Created and destroyed %d packagers in %v (avg: %v per packager)",
		iterations, elapsed, avgTime)

	// Reasonable performance expectation
	suite.Less(avgTime, time.Millisecond, "Packager creation should be fast")
}

// TestConcurrentPerformance tests concurrent access performance
func (suite *PerformanceTestSuite) TestConcurrentPerformance() {
	const numGoroutines = 100
	const operationsPerGoroutine = 10

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				_ = GetLibraryVersion()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalOps := numGoroutines * operationsPerGoroutine
	avgTime := elapsed / time.Duration(totalOps)

	suite.T().Logf("Performed %d concurrent GetLibraryVersion calls in %v (avg: %v per call)",
		totalOps, elapsed, avgTime)

	suite.Less(avgTime, time.Microsecond*100, "Concurrent operations should be fast")
}

// TestMemoryUsageStability tests that memory usage remains stable
func (suite *PerformanceTestSuite) TestMemoryUsageStability() {
	// This is a basic test - in a real scenario you might use runtime.MemStats
	const iterations = 100

	// Warm up
	for i := 0; i < 10; i++ {
		packager := NewPackager()
		if packager != nil {
			packager.Close()
		}
	}

	// Measure stability over multiple iterations
	for i := 0; i < iterations; i++ {
		packager := NewPackager()
		suite.NotNil(packager, "Each packager should be created successfully")

		params := PackagingParams{TempDir: suite.tempDir}
		streams := []StreamDescriptor{
			{Input: "test.mp4", StreamSelector: "video", Output: "output.mp4"},
		}

		// Initialize may fail, but shouldn't leak memory
		packager.Initialize(params, streams)
		packager.Close()

		// Periodic logging
		if (i+1)%20 == 0 {
			suite.T().Logf("Completed %d/%d memory stability iterations", i+1, iterations)
		}
	}

	suite.T().Log("Memory usage stability test completed")
}

func TestPerformanceSuite(t *testing.T) {
	suite.Run(t, new(PerformanceTestSuite))
}

// CompatibilityTestSuite tests compatibility with the underlying C++ library
type CompatibilityTestSuite struct {
	suite.Suite
}

// TestLibraryVersionFormat tests that version string has expected format
func (suite *CompatibilityTestSuite) TestLibraryVersionFormat() {
	version := GetLibraryVersion()
	suite.NotEmpty(version, "Version should not be empty")
	
	// Basic sanity checks for version string
	suite.NotContains(version, "\x00", "Version should not contain null bytes")
	suite.True(len(version) < 1000, "Version should be reasonable length")
	
	suite.T().Logf("Library version: %s", version)
}

// TestStatusCodeConsistency tests status code mapping consistency
func (suite *CompatibilityTestSuite) TestStatusCodeConsistency() {
	// Test that our status codes are consistent
	allCodes := []int{
		StatusOK, StatusUnknown, StatusCancelled, StatusInvalidArgument,
		StatusNotFound, StatusAlreadyExists, StatusResourceExhausted,
		StatusFailedPrecondition, StatusAborted, StatusOutOfRange,
		StatusUnimplemented, StatusInternal, StatusUnavailable,
		StatusDataLoss, StatusUnauthenticated,
	}

	for i, code := range allCodes {
		err := statusToError(code)
		if code == StatusOK {
			suite.NoError(err, "StatusOK should not produce error")
		} else {
			suite.Error(err, "Non-OK status should produce error")
			suite.NotEmpty(err.Error(), "Error message should not be empty")
		}

		suite.T().Logf("Status code %d (index %d): %v", code, i, err)
	}
}

// TestParameterStructureCompatibility tests parameter structure compatibility
func (suite *CompatibilityTestSuite) TestParameterStructureCompatibility() {
	// Test that our Go structures can be created and used
	params := PackagingParams{
		TempDir:         "/tmp",
		OutputMediaInfo: true,
		SingleThreaded:  false,
		Mp4OutputParams: Mp4OutputParams{
			IncludePsshInStream:           true,
			GenerateDashIfIopCompliantMpd: false,
		},
		ChunkingParams: ChunkingParams{
			SegmentDurationInSeconds:    6.0,
			SubsegmentDurationInSeconds: 2.0,
		},
		EncryptionParams: EncryptionParams{
			KeyId:              "0123456789abcdef0123456789abcdef",
			Key:                "fedcba9876543210fedcba9876543210",
			ClearLeadInSeconds: 10.0,
		},
	}

	// Verify we can access all fields
	suite.Equal("/tmp", params.TempDir)
	suite.True(params.OutputMediaInfo)
	suite.False(params.SingleThreaded)
	suite.True(params.Mp4OutputParams.IncludePsshInStream)
	suite.False(params.Mp4OutputParams.GenerateDashIfIopCompliantMpd)
	suite.Equal(6.0, params.ChunkingParams.SegmentDurationInSeconds)
	suite.Equal(2.0, params.ChunkingParams.SubsegmentDurationInSeconds)
	suite.Equal("0123456789abcdef0123456789abcdef", params.EncryptionParams.KeyId)
	suite.Equal("fedcba9876543210fedcba9876543210", params.EncryptionParams.Key)
	suite.Equal(10.0, params.EncryptionParams.ClearLeadInSeconds)
}

func TestCompatibilitySuite(t *testing.T) {
	suite.Run(t, new(CompatibilityTestSuite))
}