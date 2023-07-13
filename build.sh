#!/bin/sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
pushd $SCRIPT_DIR > /dev/null

mkdir -p ./build

echo "Building for linux..."
cd cmd/decoder
GOOS=linux GOARCH=amd64 go build
cd ../../cmd/encoder
GOOS=linux GOARCH=amd64 go build
cd ../..
mv cmd/decoder/decoder ./build/decoder-linux-amd64
mv cmd/encoder/encoder ./build/encoder-linux-amd64

echo "Building for Windows..."
cd cmd/decoder
GOOS=windows GOARCH=amd64 go build
cd ../../cmd/encoder
GOOS=windows GOARCH=amd64 go build
cd ../..
mv cmd/decoder/decoder.exe ./build/decoder-windows-amd64.exe
mv cmd/encoder/encoder.exe ./build/encoder-windows-amd64.exe

echo "Building for macOS..."
cd cmd/decoder
GOOS=darwin GOARCH=amd64 go build
cd ../../cmd/encoder
GOOS=darwin GOARCH=amd64 go build
cd ../..
mv cmd/decoder/decoder ./build/decoder-darwin-amd64
mv cmd/encoder/encoder ./build/encoder-darwin-amd64

popd > /dev/null
