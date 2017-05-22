// +build ignore

package main

import (
	"flag"
	"fmt"

	"github.com/cfreeman/embd"

	_ "github.com/cfreeman/embd/host/all"
)

func main() {
	flag.Parse()

	embd.InitGPIO()
	defer embd.CloseGPIO()

	val, _ := embd.AnalogRead(0)
	fmt.Printf("Reading: %v\n", val)
}
