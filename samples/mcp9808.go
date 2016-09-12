package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/kidoman/embd/sensor/mcp9808"
	"github.com/stianeikeland/go-rpio"
)

func main() {
	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	therm := mcp9808.New(bus)

	if id, err := therm.ManufacturerID(); err == nil {
		fmt.Printf("Manufacturer ID: 0x%x\n", id)
	}

	if devID, rev, err := therm.DeviceID(); err == nil {
		fmt.Printf("Device ID: 0x%x rev. 0x%x\n", devID, rev)
	}

	therm.SetShutdownMode(false)

	//therm.SetWindowTempLock(false)
	// therm.SetCriticalTempLock(false)

	config, err := therm.WriteConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("New Config: %b\n", config)
	therm.SetAlertControl(true)
	//therm.SetInterruptClear(true)
	//therm.SetAlertStatus(true)
	//therm.SetAlertSelect(false)
	//therm.SetAlertPolarity(true)
	//therm.SetAlertMode(true)
	config, err = therm.WriteConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("New Config: %b\n", config)

	if err := therm.SetWindowTempUpper(TempFToC(80)); err != nil {
		panic(err)
	}

	temp, err := therm.AmbientTemp()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Temp is %f\n", TempCToF(temp.CelsiusDeg))

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	if err := rpio.Open(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	defer rpio.Close()
	defer embd.CloseGPIO()

	alert := rpio.Pin(4)
	alert.Input()
	alert.PullDown()

	timer := time.Tick(time.Duration(5) * time.Second)

	cancel := make(chan bool)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		cancel <- true
	}()
	for {
		select {
		case <-timer:
			temp, err := therm.AmbientTemp()
			if err == nil {
				fmt.Printf("Ambient temp is: %f\n", TempCToF(temp.CelsiusDeg))
			}
			status := alert.Read()
			fmt.Printf("Status: %d\n\n", status)
			if status == rpio.High {
				fmt.Println("Alert temp has been reached.")
				return
			}
		case <-cancel:
			return
		}
	}
}

func TempCToF(tempC float64) float64 {
	return tempC*9/5 + 32
}

func TempFToC(tempF float64) float64 {
	return (tempF - 32) * 5 / 9
}
