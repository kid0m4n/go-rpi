// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/kidoman/embd"
	"github.com/kidoman/embd/sensor/sht1x"

	_ "github.com/kidoman/embd/host/all"
)

func main() {
	flag.Parse()

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	data, err := embd.NewDigitalPin(4)
	if err != nil {
		panic(err)
	}
	defer data.Close()

	clock, err := embd.NewDigitalPin(3)
	if err != nil {
		panic(err)
	}
	defer clock.Close()

	sensor := sht1x.New(data, clock)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	defer signal.Stop(quit)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m, err := sensor.Measure()
			if err != nil {
				fmt.Printf("err: %v\n", err)
			}
			fmt.Printf("Temperature: %.1fC, Relative Humidity: %.1f%%, Dew Point: %.1fC\n", m.Temperature, m.RelativeHumidity, m.DewPoint)

		case <-quit:
			return
		}
	}
}
