// +build ignore

// This sample runs on a monochrome 128x64 OLED graphic display using an SSD1306 controller,
// such as https://www.adafruit.com/product/938.
// It demonstrates the rectangular fill/clear and point on/off operations animated across the display.

package main

import (
	"github.com/golang/glog"
	"github.com/kidoman/embd"
	"github.com/kidoman/embd/controller/ssd1306"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver

	"flag"
	"os"
	"os/signal"
	"time"
)

func main() {
	flag.Parse()
	glog.Info("Starting")

	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()

	spiBus := embd.NewSPIBus(embd.SPIMode0, 0, 1000000, 8, 0)
	defer spiBus.Close()

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	dcPin := setupPin("GPIO_23")
	defer dcPin.Close()
	resetPin := setupPin("GPIO_24")
	defer resetPin.Close()

	controller, err := ssd1306.NewSPI(spiBus, dcPin, resetPin, 128, 64)
	if err != nil {
		glog.Fatalf("Failed to start: %s", err)
	}
	defer controller.Close()

	buffer := controller.NewBuffer()
	first := true
	chunkWidth := 32
	chunkHeight := 20
	zipDir := 1
	zipX := 0
	prevZipX := 0
	var prevX, prevY int

	// Setup Control-C to gracefully stop
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

outer:
	for {

		for y := 0; y+chunkHeight <= 64; y += chunkHeight {
			for x := 0; x+chunkWidth <= 128; x += chunkWidth {
				select {
				case <-done:
					break outer
				default:
					glog.Infof("x=%d, y=%d\n", x, y)
					if !first {
						buffer.ClearRect(prevX, prevY, chunkWidth, chunkHeight)
					} else {
						first = false
					}
					buffer.FillRect(x, y, chunkWidth, chunkHeight)
					prevX = x
					prevY = y

					buffer.Off(prevZipX, 63)
					buffer.On(zipX, 63)
					prevZipX = zipX
					zipX += zipDir
					if zipX >= 128 {
						zipX = 127
						zipDir = -1
					} else if zipX < 0 {
						zipX = 0
						zipDir = 1
					}

					controller.Display(buffer)

					time.Sleep(100 * time.Millisecond)
				}
			}
		}

	}

}

func setupPin(key string) embd.DigitalPin {

	p, err := embd.NewDigitalPin(key)
	if err != nil {
		panic(err)
	}

	if err := p.SetDirection(embd.Out); err != nil {
		panic(err)
	}

	return p
}
