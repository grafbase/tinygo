package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/goenv"
)

func TestWasiPackageEmbed(t *testing.T) {
	// Skip if wasm-tools is not installed
	wasmTools := goenv.Get("WASMTOOLS")
	if wasmTools == "" {
		t.Skip("WASMTOOLS environment variable not set, skipping test")
	}

	_, err := exec.LookPath(wasmTools)
	if err != nil {
		t.Skipf("wasm-tools not found at %s, skipping test", wasmTools)
	}

	// Skip if wasmtime is not installed (needed for verification)
	_, err = exec.LookPath("wasmtime")
	if err != nil {
		t.Skip("wasmtime not found in PATH, skipping test")
	}

	tinygoRoot := goenv.Get("TINYGOROOT")
	if tinygoRoot == "" {
		t.Fatal("TINYGOROOT environment variable not set")
	}

	// Test file
	testFile := filepath.Join("testdata", "wasi-package-test", "custom_wasi_test.go")

	tests := []struct {
		name             string
		wasiPackage      string
		witPackage       string
		witWorld         string
		expectedInWat    []string
		notExpectedInWat []string
	}{
		{
			name: "Default WASI CLI",
			// No custom WASI package - should use default from wasip2.json
			witPackage: filepath.Join(tinygoRoot, "lib", "wasi-cli", "wit"),
			witWorld:   "wasi:cli/command",
			expectedInWat: []string{
				"(import \"wasi:cli/command\" \"run\"",
				"(export \"hello\"",
				"(export \"add\"",
			},
			// Default should NOT include our custom world
			notExpectedInWat: []string{
				"(import \"test:custom-wasi/hello\"",
				"(export \"test:custom-wasi/hello\"",
			},
		},
		{
			name:        "Custom WASI World",
			wasiPackage: filepath.Join("testdata", "wasi-package-test", "wit"),
			witPackage:  filepath.Join("testdata", "wasi-package-test", "wit"),
			witWorld:    "test:custom-wasi/custom-world",
			expectedInWat: []string{
				"(export \"test:custom-wasi/hello\"",
				"(export \"add\"",
				"(export \"get-greeting\"",
			},
			// Custom world should NOT include the standard wasi:cli imports
			notExpectedInWat: []string{
				"(import \"wasi:cli/command\" \"run\"",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputWasm := filepath.Join(tmpDir, "output.wasm")
			watFile := filepath.Join(tmpDir, "output.wat")

			// Build the test file with the specified configuration
			args := []string{
				"build",
				"-target=wasip2",
				"-o", outputWasm,
			}

			if tc.wasiPackage != "" {
				args = append(args, "-wasi-package="+tc.wasiPackage)
			}
			if tc.witPackage != "" {
				args = append(args, "-wit-package="+tc.witPackage)
			}
			if tc.witWorld != "" {
				args = append(args, "-wit-world="+tc.witWorld)
			}

			args = append(args, testFile)

			// Run TinyGo build
			cmd := exec.Command("./tinygo", args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Failed to build test file: %v\nStderr: %s", err, stderr.String())
			}

			// Verify the WASM file exists
			if _, err := os.Stat(outputWasm); os.IsNotExist(err) {
				t.Fatalf("Output WASM file was not created: %s", outputWasm)
			}

			// Convert WASM to WAT for inspection
			cmd = exec.Command(wasmTools, "print", outputWasm)
			var watOutput bytes.Buffer
			cmd.Stdout = &watOutput
			cmd.Stderr = os.Stderr

			err = cmd.Run()
			if err != nil {
				t.Fatalf("Failed to convert WASM to WAT: %v", err)
			}

			// Save WAT content to file for debugging
			err = os.WriteFile(watFile, watOutput.Bytes(), 0644)
			if err != nil {
				t.Fatalf("Failed to write WAT file: %v", err)
			}

			watContent := watOutput.String()

			// Check for expected patterns in the WAT output
			for _, expected := range tc.expectedInWat {
				if !strings.Contains(watContent, expected) {
					t.Errorf("Expected WAT to contain: %s", expected)
					t.Logf("WAT file available at: %s", watFile)
				}
			}

			// Check for patterns that should NOT be in the WAT output
			for _, notExpected := range tc.notExpectedInWat {
				if strings.Contains(watContent, notExpected) {
					t.Errorf("WAT should NOT contain: %s", notExpected)
					t.Logf("WAT file available at: %s", watFile)
				}
			}

			// Validate with wasmtime that the component is valid
			cmd = exec.Command("wasmtime", "validate", "--wasm", "component-model", outputWasm)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Component validation failed: %v\nOutput: %s", err, string(output))
			}
		})
	}
}

// TestWasiPackageEmbedErrors tests error cases for WASI package embedding
func TestWasiPackageEmbedErrors(t *testing.T) {
	// Skip if wasm-tools is not installed
	wasmTools := goenv.Get("WASMTOOLS")
	if wasmTools == "" {
		t.Skip("WASMTOOLS environment variable not set, skipping test")
	}

	_, err := exec.LookPath(wasmTools)
	if err != nil {
		t.Skipf("wasm-tools not found at %s, skipping test", wasmTools)
	}

	// Test file
	testFile := filepath.Join("testdata", "wasi-package-test", "custom_wasi_test.go")

	tests := []struct {
		name          string
		wasiPackage   string
		witPackage    string
		witWorld      string
		expectedError string
	}{
		{
			name:          "Invalid WASI Package Path",
			wasiPackage:   "/nonexistent/path",
			witPackage:    "/nonexistent/path",
			witWorld:      "test:custom-wasi/custom-world",
			expectedError: "wasm-tools component embed.*failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputWasm := filepath.Join(tmpDir, "output.wasm")

			// Build the test file with the specified configuration
			args := []string{
				"build",
				"-target=wasip2",
				"-o", outputWasm,
			}

			if tc.wasiPackage != "" {
				args = append(args, "-wasi-package="+tc.wasiPackage)
			}
			if tc.witPackage != "" {
				args = append(args, "-wit-package="+tc.witPackage)
			}
			if tc.witWorld != "" {
				args = append(args, "-wit-world="+tc.witWorld)
			}

			args = append(args, testFile)

			// Run TinyGo build and expect failure
			cmd := exec.Command("./tinygo", args...)
			output, err := cmd.CombinedOutput()

			// Should fail
			if err == nil {
				t.Fatalf("Expected build to fail but it succeeded")
			}

			// Check for expected error message
			re := regexp.MustCompile(tc.expectedError)
			if !re.Match(output) {
				t.Errorf("Expected error matching '%s', got: %s", tc.expectedError, string(output))
			}
		})
	}
}
