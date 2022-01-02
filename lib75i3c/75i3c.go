package lib75i3c

import (
	"bytes"
	"image"
	"image/color"
	"time"

	"github.com/stianeikeland/go-rpio"
)

const (
	width  = 880
	height = 528

	X = width
	Y = height

	resetPin = rpio.Pin(17)
	dcPin    = rpio.Pin(25)
	csPin    = rpio.Pin(8)
	busyPin  = rpio.Pin(24)
)

const (
	SOFTSTARTSETTING byte = 0x0C
	DEEPSLEEP        byte = 0x10
	DATAENTRYMODE    byte = 0x11
	SWRESET          byte = 0x12
	WRITEBLACK       byte = 0x24
	WRITERED         byte = 0x26
	VBD              byte = 0x3C
)

// Internal usage of width and height
var intW = w()
var intH = height

// w calculates internal representation of width
func w() int {
	if width%8 == 0 {
		return width / 8
	}
	return width/8 + 1
}
func init() {
	if err := rpio.Open(); err != nil {
		panic(err)
	}

	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		panic(err)
	}

	resetPin.Mode(rpio.Output)
	dcPin.Mode(rpio.Output)
	csPin.Mode(rpio.Output)
	busyPin.Mode(rpio.Input)
	rpio.SpiSpeed(4000000)
	rpio.SpiMode(0, 0)
}

// Exit closes the SPI interface to the display
func Exit() {
	rpio.SpiEnd(rpio.Spi0)
	csPin.Write(rpio.Low)
	dcPin.Write(rpio.Low)
	resetPin.Write(rpio.Low)
	rpio.Close()
}

func Reset() {
	resetPin.Write(rpio.High)
	time.Sleep(200 * time.Millisecond)
	resetPin.Write(rpio.Low)
	time.Sleep(2 * time.Millisecond)
	resetPin.Write(rpio.High)
	time.Sleep(200 * time.Millisecond)
}

func SendCommand(reg byte) {
	dcPin.Write(rpio.Low)
	csPin.Write(rpio.Low)
	rpio.SpiTransmit(reg)
	csPin.Write(rpio.High)
}

func SendData(data byte) {
	dcPin.Write(rpio.High)
	csPin.Write(rpio.Low)
	rpio.SpiTransmit(data)
	csPin.Write(rpio.High)
}

func WaitIdle() {
	for busyPin.Read() == rpio.High {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
}

func TurnOnDisplay() {
	SendCommand(0x22)
	SendData(0xC7) // Load LUT from MCU (0x32)
	SendCommand(0x20)
	time.Sleep(200 * time.Millisecond) // The delay is necessary, at least 200 us
	WaitIdle()
}

func Initialize() {
	Reset()
	SendCommand(SWRESET)
	WaitIdle()

	SendCommand(0x46) // Auto Write RAM
	SendData(0xF7)
	WaitIdle()

	SendCommand(0x47) // Auto Write RAM
	SendData(0xF7)
	WaitIdle()

	SendCommand(SOFTSTARTSETTING)
	SendData(0xAE)
	SendData(0xC7)
	SendData(0xC3)
	SendData(0xC0)
	SendData(0x40)

	SendCommand(0x01) // Set MUX as 527
	SendData(0xAF)
	SendData(0x02)
	SendData(0x01)

	SendCommand(DATAENTRYMODE)
	SendData(0x01)

	SendCommand(0x44)
	SendData(0x00) // Ram x address start at 0
	SendData(0x00)
	SendData(0x6F) // Ram x address start at 36Fh -> 879
	SendData(0x03)

	SendCommand(0x45)
	SendData(0xAF) // Ram y address starts at 20Fh
	SendData(0x02)
	SendData(0x00) /// Ram y address ends at 00h
	SendData(0x00)

	SendCommand(VBD)
	SendData(0x01) // LUT1, white

	SendCommand(0x18)
	SendData(0x80)
	SendCommand(0x22)
	SendData(0xB1) // Load temperature and waveform setting
	SendCommand(0x20)
	WaitIdle() // Wait for epaper IC to release idle signal

	SendCommand(0x4E)
	SendData(0x00)
	SendData(0x00)

	prepDisplayDraw()
}

func ClearDisplay() {
	prepDisplayDraw()

	SendCommand(WRITEBLACK)
	for i := 0; i < intW*intH; i++ {
		SendData(0xFF)
	}

	SendCommand(WRITERED)
	for i := 0; i < intW*intH; i++ {
		SendData(0x00)
	}
	TurnOnDisplay()
}

func prepDisplayDraw() {
	SendCommand(0x4F)
	SendData(0xAF)
	SendData(0x02)
}

// DisplayImage displays the bitmap in `black` with black pixels and the bitmap in `red` as red pixels
func DisplayImage(black image.Image, red image.Image) {
	prepDisplayDraw()
	defer TurnOnDisplay()

	if black != nil {
		img := Convert(black)
		SendCommand(WRITEBLACK)
		for j := 0; j < intH; j++ {
			for i := 0; i < intW; i++ {
				SendData(img[i+j*intW])
			}
		}
	}

	if red != nil {
		SendCommand(WRITERED)
		img := Convert(red)
		for j := 0; j < intH; j++ {
			for i := 0; i < intW; i++ {
				SendData(^img[i+j*intW])
			}
		}
	}
}

func Sleep() {
	SendCommand(DEEPSLEEP)
	SendData(0x01)
}

// Convert converts the input image into a ready-to-display byte buffer.
func Convert(img image.Image) []byte {
	var byteToSend byte = 0x00
	var bgColor = 1

	buffer := bytes.Repeat([]byte{0x00}, intW*intH)

	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			bit := bgColor

			if i < img.Bounds().Dx() && j < img.Bounds().Dy() {
				bit = color.Palette([]color.Color{color.Black, color.White}).Index(img.At(i, j))
			}

			if bit == 1 {
				//byteToSend |= 0x80 >> (uint32(i) % 8)
				byteToSend |= 0xFF >> (uint32(i) % 8)
			}

			if i%8 == 7 {
				buffer[(i/8)+(j*intW)] = byteToSend
				byteToSend = 0x00
			}
		}
	}
	return buffer
}
