// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
	"github.com/kidoman/embd/sensor/isl29125"
)

func main() {

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	isl := isl29125.New(isl29125.DefaultConfig, bus)
	defer isl.Close()
	isl.Init()

	for {
		r, err := isl.Reading()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v", r)

		time.Sleep(500 * time.Millisecond)
	}
}
