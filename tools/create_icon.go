package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

// Simple icon generator - creates a computer/port-scanner themed icon
func main() {
	// Create a 32x32 image
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	
	// Fill background with dark blue
	dark := color.RGBA{30, 30, 60, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{dark}, image.Point{}, draw.Src)
	
	// Draw a simple computer/monitor shape
	light := color.RGBA{100, 150, 255, 255}
	
	// Monitor frame
	for x := 6; x < 26; x++ {
		for y := 8; y < 20; y++ {
			if x == 6 || x == 25 || y == 8 || y == 19 {
				img.Set(x, y, light)
			} else if x > 8 && x < 23 && y > 10 && y < 17 {
				// Screen area
				img.Set(x, y, color.RGBA{50, 200, 100, 255})
			}
		}
	}
	
	// Monitor stand
	for x := 14; x < 18; x++ {
		for y := 20; y < 24; y++ {
			img.Set(x, y, light)
		}
	}
	
	// Base
	for x := 12; x < 20; x++ {
		img.Set(x, 24, light)
	}
	
	// Add some "port" dots
	green := color.RGBA{0, 255, 0, 255}
	red := color.RGBA{255, 0, 0, 255}
	
	img.Set(10, 13, green)
	img.Set(12, 13, green) 
	img.Set(14, 13, red)
	img.Set(16, 13, green)
	img.Set(18, 13, red)
	img.Set(20, 13, green)
	
	// Save as PNG first (we'll convert to ICO later)
	var buf bytes.Buffer
	png.Encode(&buf, img)
	
	file, err := os.Create("icon.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	
	file.Write(buf.Bytes())
	println("Created icon.png - now convert to .ico using online tool or ImageMagick")
}