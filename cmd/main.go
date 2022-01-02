package main

import (
	"time"

	"github.com/adamhassel/epaper/lib75i3c"
	"github.com/fogleman/gg"
)

func main() {
	mine()
}

func mine() {
	im, err := gg.LoadPNG("screenshot.png")
	if err != nil {
		panic(err)
	}

	defer lib75i3c.Exit()
	lib75i3c.Initialize()
	lib75i3c.ClearDisplay()
	time.Sleep(500 * time.Millisecond)
	lib75i3c.DisplayImage(im, nil)
	time.Sleep(500 * time.Millisecond)
	lib75i3c.Sleep()
}
