#!/bin/bash

# Helper script for converting WASM files to WAT for inspection
# Usage: ./wasm2wat.sh input.wasm [output.wat]

set -e

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 input.wasm [output.wat]"
  exit 1
fi

INPUT="$1"
OUTPUT="${2:-${INPUT%.wasm}.wat}"

# Try to use wasm-tools first, fall back to wasm2wat from WABT if available
if command -v wasm-tools &> /dev/null; then
  echo "Using wasm-tools to convert $INPUT to $OUTPUT"
  wasm-tools print "$INPUT" > "$OUTPUT"
elif command -v wasm2wat &> /dev/null; then
  echo "Using wasm2wat to convert $INPUT to $OUTPUT"
  wasm2wat "$INPUT" -o "$OUTPUT"
else
  echo "Error: Neither wasm-tools nor wasm2wat (WABT) found in PATH"
  echo "Please install one of these tools to convert WASM to WAT"
  exit 1
fi

echo "Converted $INPUT to $OUTPUT"
