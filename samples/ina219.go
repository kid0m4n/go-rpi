// +build ignore

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/kidoman/embd"
	"github.com/kidoman/embd/sensor/ina219"

	_ "github.com/kidoman/embd/host/all"
)

func main() {
	flag.Parse()

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	ina := ina219.New(bus, 0x40, 0.001)
	defer ina.Close()

	for {
		sv, err := ina.ShuntVoltage()
		if err != nil {
			panic(err)
		}

		v, err := ina.Voltage()
		if err != nil {
			panic(err)
		}

		c, err := ina.Current()
		if err != nil {
			panic(err)
		}

		p, err := ina.Power()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Shunt Voltage=%v  Voltage=%v Current=%v Power=%v \n", sv, v, c, p)

		time.Sleep(500 * time.Millisecond)
	}
}
