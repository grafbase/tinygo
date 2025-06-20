# WASI Package Test

This directory contains test files for the custom WASI package embedding functionality in TinyGo.

## Overview

TinyGo supports building WebAssembly components that implement various WASI worlds. By default, TinyGo's `wasip2` target uses the standard `wasi:cli/command` world, but with the new `-wasi-package` flag, you can now specify a custom WASI world to implement.

## Files in this Directory

- `custom_wasi_test.go`: A simple Go program that exports WebAssembly functions
- `wit/custom-world.wit`: A minimal WIT definition for a custom WASI world
- `wasm2wat.sh`: Helper script to convert WASM to WAT for inspection

## Usage Example

### Building with Default WASI CLI World

```bash
tinygo build -target=wasip2 -o output.wasm custom_wasi_test.go
```

This will build a WebAssembly component that implements the default `wasi:cli/command` world.

### Building with Custom WASI World

```bash
tinygo build -target=wasip2 \
  -wasi-package=./wit \
  -wit-package=./wit \
  -wit-world=test:custom-wasi/custom-world \
  -o output.wasm custom_wasi_test.go
```

This will build a WebAssembly component that implements the custom world defined in `wit/custom-world.wit`.

## Running the Components

You can run the built components using Wasmtime with the component model enabled:

```bash
# For default WASI CLI world
wasmtime run --wasm component-model output.wasm

# For custom WASI world, you need a JavaScript adapter
echo 'import { hello } from "test:custom-wasi/hello"; console.log(hello.getGreeting());' > adapter.js
wasmtime run --wasm component-model --enable-component-model output.wasm adapter.js
```

## Testing

The tests for this feature are in:

1. `tinygo/wasi_package_test.go` - Tests that verify the correct structure of the generated WebAssembly components
2. `tinygo/wasi_package_integration_test.go` - Tests that verify the components can be executed correctly

Run the tests with:

```bash
cd tinygo
go test -v -run TestWasiPackage
```
