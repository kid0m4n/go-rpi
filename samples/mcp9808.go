package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/kidoman/embd/sensor/mcp9808"
)

func main() {
	bus := embd.NewI2CBus(1)
	defer embd.CloseI2C()

	therm := mcp9808.New(bus)
	// set sensor to low power mode when we're done
	defer therm.SetShutdownMode(true)

	if id, err := therm.ManufacturerID(); err == nil {
		fmt.Printf("Manufacturer ID: 0x%x\n", id)
	}

	if devID, rev, err := therm.DeviceID(); err == nil {
		fmt.Printf("Device ID: 0x%x rev. 0x%x\n", devID, rev)
	}

	therm.SetShutdownMode(false)
	therm.SetAlertMode(true)
	therm.SetInterruptClear(true)
	therm.SetAlertStatus(true)
	therm.SetAlertControl(true)
	therm.SetAlertSelect(false)
	therm.SetAlertPolarity(true)

	config, _ := therm.Config()
	fmt.Printf("New Config: %b\n", config)

	if err := therm.SetCriticalTemp(TempFToC(90)); err != nil {
		panic(err)
	}

	if err := therm.SetWindowTempLower(TempFToC(-40)); err != nil {
		panic(err)
	}
	lowerTemp, err := therm.WindowTempLower()
	if err != nil {
		fmt.Printf("Error reading lower temp limit: %s\n", err.Error())
	}
	fmt.Printf("Lower Temp Limit set to: %fC\n", lowerTemp)

	if err := therm.SetWindowTempUpper(TempFToC(80)); err != nil {
		panic(err)
	}
	upperTemp, _ := therm.WindowTempUpper()
	fmt.Printf("Upper Temp Limit set to: %fC\n", upperTemp)

	alert, err := embd.NewDigitalPin(23)
	if err != nil {
		panic(err)
	}
	defer alert.Close()

	alert.SetDirection(embd.In)
	alert.ActiveLow(false)

	err = alert.Watch(embd.EdgeRising, func(alert embd.DigitalPin) {
		fmt.Printf("Temperature is outside the specified window!\n")
		therm.SetInterruptClear(true)
		therm.Config()
	})
	if err != nil {
		panic(err)
	}

	cancel := make(chan bool)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		cancel <- true
	}()

	timer := time.Tick(time.Duration(5) * time.Second)
	for {
		select {
		case <-timer:
			temp, err := therm.AmbientTemp()
			if err != nil {
				fmt.Printf("Error reading temp: %s\n", err.Error())
			} else {
				fmt.Printf("Current temp is: %fF (%fC), Window Alert: %v, Critical Alert: %v\n",
					TempCToF(temp.CelsiusDeg), temp.CelsiusDeg, temp.AboveUpper || temp.BelowLower, temp.AboveCritical)
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
