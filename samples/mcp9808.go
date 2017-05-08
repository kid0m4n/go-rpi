package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/kidoman/embd/sensor/mcp9808"
)

func main() {
	bus := embd.NewI2CBus(1)
	defer embd.CloseI2C()

	therm, _ := mcp9808.New(bus)
	// set sensor to low power mode when we're done
	defer therm.SetShutdownMode(true)

	if id, err := therm.ManufacturerID(); err == nil {
		fmt.Printf("Manufacturer ID: 0x%x\n", id)
	}

	if devID, rev, err := therm.DeviceID(); err == nil {
		fmt.Printf("Device ID: 0x%x rev. 0x%x\n", devID, rev)
	}

	therm.SetShutdownMode(false)
	therm.SetCriticalTempLock(false)
	therm.SetWindowTempLock(false)
	therm.SetAlertMode(true)
	therm.SetInterruptClear(true)
	therm.SetAlertStatus(true)
	therm.SetAlertControl(true)
	therm.SetAlertSelect(false)
	therm.SetAlertPolarity(true)

	// get faster results (130ms vs 250ms default)
	therm.SetTempResolution(mcp9808.EighthC)

	config, _ := therm.Config()
	fmt.Printf("New Config: %b\n", config)

	if err := therm.SetCriticalTemp(TempFToC(95)); err != nil {
		panic(err)
	}
	critTemp, err := therm.CriticalTemp()
	if err != nil {
		fmt.Printf("Error reading critical temp limit: %s\n", err.Error())
	}
	fmt.Printf("Critical Temp set to: %fC\n", critTemp)

	if err := therm.SetWindowTempLower(TempFToC(60)); err != nil {
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

	therm.SetCriticalTempLock(true)
	therm.SetWindowTempLock(true)

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

	for {
		select {
		case <-cancel:
			return
		default:
			temp, err := therm.AmbientTemp()
			if err != nil {
				fmt.Printf("Error reading temp: %s\n", err.Error())
			} else {
				fmt.Printf("Current temp is: %fF (%fC), Window Alert: %v, Critical Alert: %v\n",
					TempCToF(temp.CelsiusDeg), temp.CelsiusDeg, temp.AboveUpper || temp.BelowLower, temp.AboveCritical)
			}
		}
	}
}

func TempCToF(tempC float64) float64 {
	return tempC*9/5 + 32
}

func TempFToC(tempF float64) float64 {
	return (tempF - 32) * 5 / 9
}
