package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/adamhassel/epaper/lib75i3c"
)

func main() {
	var black, red image.Image
	var err error
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s blackimage.png [redimage.png]\n", os.Args[0])
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		black, err = LoadPNG(os.Args[1])
		if err != nil {
			panic(err)
		}
	}
	if len(os.Args) > 2 {
		red, err = LoadPNG(os.Args[2])
		if err != nil {
			panic(err)
		}
	}

	defer lib75i3c.Exit()
	lib75i3c.Initialize()
	lib75i3c.ClearDisplay()
	time.Sleep(500 * time.Millisecond)
	lib75i3c.DisplayImage(black, red)
	time.Sleep(500 * time.Millisecond)
	lib75i3c.Sleep()
}

func LoadPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}
