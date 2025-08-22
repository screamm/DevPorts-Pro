package main

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"os"
)

// Simple PNG to ICO converter
func main() {
	// Read the PNG file
	pngFile, err := os.Open("icon.png")
	if err != nil {
		panic(err)
	}
	defer pngFile.Close()

	// Decode PNG
	img, err := png.Decode(pngFile)
	if err != nil {
		panic(err)
	}

	// Convert back to PNG bytes for embedding
	var pngBuf bytes.Buffer
	err = png.Encode(&pngBuf, img)
	if err != nil {
		panic(err)
	}
	pngData := pngBuf.Bytes()

	// Create ICO file
	icoFile, err := os.Create("icon.ico")
	if err != nil {
		panic(err)
	}
	defer icoFile.Close()

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// ICO Header (6 bytes)
	binary.Write(icoFile, binary.LittleEndian, uint16(0))    // Reserved
	binary.Write(icoFile, binary.LittleEndian, uint16(1))    // Image type (1 = icon)
	binary.Write(icoFile, binary.LittleEndian, uint16(1))    // Number of images

	// Image Directory Entry (16 bytes)
	binary.Write(icoFile, binary.LittleEndian, uint8(width))  // Width (0 = 256)
	binary.Write(icoFile, binary.LittleEndian, uint8(height)) // Height (0 = 256) 
	binary.Write(icoFile, binary.LittleEndian, uint8(0))      // Color palette entries
	binary.Write(icoFile, binary.LittleEndian, uint8(0))      // Reserved
	binary.Write(icoFile, binary.LittleEndian, uint16(0))   // Color planes
	binary.Write(icoFile, binary.LittleEndian, uint16(32))  // Bits per pixel
	binary.Write(icoFile, binary.LittleEndian, uint32(len(pngData))) // Data size
	binary.Write(icoFile, binary.LittleEndian, uint32(22))  // Data offset

	// PNG data
	icoFile.Write(pngData)

	println("Created icon.ico")
}