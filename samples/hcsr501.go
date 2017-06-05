// +build ignore

package main

import (
	"flag"
	"log"
	"time"

	"github.com/kidoman/embd"

	_ "github.com/kidoman/embd/host/chip"
	"github.com/kidoman/embd/sensor/hcsr501"
)

func main() {
	flag.Parse()

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	trig, err := embd.NewDigitalPin(132)
	if err != nil {
		panic(err)
	}
	defer trig.Close()

	pir := hcsr501.New(trig)

	for {
		time.Sleep(3 * time.Second)
		p, err := pir.Detect()
		if err != nil {
			log.Printf("error %v", err)
			continue
		}
		log.Printf("PIR Detect %v", p)
	}
}
