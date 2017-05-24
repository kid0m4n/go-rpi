// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/kidoman/embd"
	"github.com/kidoman/embd/sensor/mpu6050"

	_ "github.com/kidoman/embd/host/all"
)

func main() {
	flag.Parse()

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	sensor, _ := mpu6050.New(bus, &mpu6050.Config{GiroScale: "1000", AccelScale: "4g", Dlpf: "6"})
	// GiroScale, Dlpf, AccelScale can be omited to use the defaults

	sensor.Start()
	defer sensor.Close()

	// catch ctrl+c and ctrl+x so the sensor.Close is executed
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	for {
		reading := sensor.Read()
		o := reading.Orientation()
		v := reading.Velocity()
		t := reading.Temperature()

		select {
		case <-stop:
			return
		default:
			print("\033[H\033[2J") // clear the screen for every read
			fmt.Println(o)
			fmt.Println(v)
			fmt.Println(t)
			time.Sleep(200 * time.Millisecond)
		}
	}
}
