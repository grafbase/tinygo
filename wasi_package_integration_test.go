package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/goenv"
)

// TestWasiPackageIntegration tests that custom WASI components can be built and executed
func TestWasiPackageIntegration(t *testing.T) {
	// Skip if wasm-tools is not installed
	wasmTools := goenv.Get("WASMTOOLS")
	if wasmTools == "" {
		t.Skip("WASMTOOLS environment variable not set, skipping test")
	}

	_, err := exec.LookPath(wasmTools)
	if err != nil {
		t.Skipf("wasm-tools not found at %s, skipping test", wasmTools)
	}

	// Skip if wasmtime is not installed
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
		name        string
		wasiPackage string
		witPackage  string
		witWorld    string
		jsAdapter   string // JavaScript adapter content to create for testing
		expected    string // Expected output when run with wasmtime
	}{
		{
			name: "Default WASI CLI Component",
			// No custom WASI package - should use default from wasip2.json
			witPackage: filepath.Join(tinygoRoot, "lib", "wasi-cli", "wit"),
			witWorld:   "wasi:cli/command",
			jsAdapter: `
				import { command } from "wasi:cli/command";
				command.run().then(() => {});
			`,
			expected: "Hello from custom WASI world!", // Output of exported hello function
		},
		{
			name:        "Custom WASI World Component",
			wasiPackage: filepath.Join("testdata", "wasi-package-test", "wit"),
			witPackage:  filepath.Join("testdata", "wasi-package-test", "wit"),
			witWorld:    "test:custom-wasi/custom-world",
			jsAdapter: `
				// Import the custom-world component
				import { hello } from "test:custom-wasi/hello";

				// Call the exported functions and print results
				console.log(hello.getGreeting());
				console.log("Addition result: " + hello.add(40, 2));
			`,
			expected: "Hello from custom WASI world!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputWasm := filepath.Join(tmpDir, "output.wasm")
			adapterFile := filepath.Join(tmpDir, "adapter.js")

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

			// Write the JS adapter file for testing
			err = os.WriteFile(adapterFile, []byte(tc.jsAdapter), 0644)
			if err != nil {
				t.Fatalf("Failed to write adapter file: %v", err)
			}

			// Run the component with wasmtime
			cmd = exec.Command("wasmtime", "run", "--wasm", "component-model",
				"--enable-component-model", "-S", "inherit-stderr", "-S", "inherit-stdout",
				"--component", outputWasm, adapterFile)

			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()
			if err != nil {
				t.Fatalf("Failed to run component with wasmtime: %v", err)
			}

			// Check output
			output := stdout.String()
			if !strings.Contains(output, tc.expected) {
				t.Errorf("Expected output to contain %q, got: %s", tc.expected, output)
			}
		})
	}
}

// TestWasiPackageInterop tests interoperability between Go code and different WASI worlds
func TestWasiPackageInterop(t *testing.T) {
	// This is a more advanced test that could be implemented in the future
	// It would test interoperability between Go code and various WASI interfaces
	// For now, we're implementing a simpler integration test above
	t.Skip("Advanced interoperability test not implemented yet")
}
