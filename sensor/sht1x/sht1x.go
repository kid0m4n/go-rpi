// Package sht1x allows interfacing with Sensirion SHT1x family of humidity
// and temperature sensors.
package sht1x

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/kidoman/embd"
)

// Constants and implementation derived from
// http://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/Humidity/Sensirion_Humidity_SHT1x_Datasheet_V5.pdf
const (
	d1                        = -40.1
	d2                        = 0.01
	c1                        = -2.0468
	c2                        = 0.0367
	c3                        = -0.0000015955
	t1                        = 0.01
	t2                        = 0.00008
	tnA0                      = 243.12
	mA0                       = 17.62
	tnB0                      = 272.62
	mB0                       = 22.46
	measureTemperatureCommand = 3
	measureHumidityCommand    = 5
)

// Measurement represents a measurement of temperature in degrees Celsius,
// relative humidity in percent, and dew point in degrees Celsius.
type Measurement struct {
	Temperature      float64
	RelativeHumidity float64
	DewPoint         float64
}

// SHT1x represents a SHT1x humidity and temperature sensor.
type SHT1x struct {
	DataPin, ClockPin embd.DigitalPin
	m                 sync.RWMutex
}

// New returns a handle to a SHT1x sensor.
func New(dataPin, clockPin embd.DigitalPin) *SHT1x {
	return &SHT1x{
		DataPin:  dataPin,
		ClockPin: clockPin,
	}
}

// Measure returns a Measurement with temperature in degrees Celsius,
// relative humidity in percent, and dew point in degrees Celsius.
func (d *SHT1x) Measure() (*Measurement, error) {
	rawTemperature, err := d.rawTemperature()
	if err != nil {
		return nil, err
	}

	rawHumidity, err := d.rawHumidity()
	if err != nil {
		return nil, err
	}

	temperature := temperatureFromRaw(rawTemperature)
	relativeHumidity := humidityFromRaw(rawHumidity, temperature)
	dewPoint := calculateDewPoint(relativeHumidity, temperature)

	measurement := &Measurement{
		Temperature:      temperature,
		RelativeHumidity: relativeHumidity,
		DewPoint:         dewPoint,
	}

	return measurement, nil
}

// Temperature returns a temperature measurement in degrees Celsius.
func (d *SHT1x) Temperature() (float64, error) {
	raw, err := d.rawTemperature()
	if err != nil {
		return 0, err
	}

	temperature := temperatureFromRaw(raw)

	return temperature, nil
}

// RelativeHumidity returns a relative humidity measurement in percent.
func (d *SHT1x) RelativeHumidity() (float64, error) {
	rawTemperature, err := d.rawTemperature()
	if err != nil {
		return 0, err
	}

	rawHumidity, err := d.rawHumidity()
	if err != nil {
		return 0, err
	}

	temperature := temperatureFromRaw(rawTemperature)
	relativeHumidity := humidityFromRaw(rawHumidity, temperature)

	return relativeHumidity, nil
}

// DewPoint returns a dew point measurement in degrees Celsius.
func (d *SHT1x) DewPoint() (float64, error) {
	rawTemperature, err := d.rawTemperature()
	if err != nil {
		return 0, err
	}

	rawHumidity, err := d.rawHumidity()
	if err != nil {
		return 0, err
	}

	temperature := temperatureFromRaw(rawTemperature)
	relativeHumidity := humidityFromRaw(rawHumidity, temperature)

	dewPoint := calculateDewPoint(relativeHumidity, temperature)
	return dewPoint, nil
}

func (d *SHT1x) sendCommand(command int) error {
	d.DataPin.SetDirection(embd.Out)
	d.ClockPin.SetDirection(embd.Out)

	d.DataPin.Write(embd.High)
	d.ClockPin.Write(embd.High)
	d.DataPin.Write(embd.Low)
	d.ClockPin.Write(embd.Low)
	d.ClockPin.Write(embd.High)
	d.DataPin.Write(embd.High)
	d.ClockPin.Write(embd.Low)

	for i := 0; i < 8; i++ {
		bitSet := (command&(1<<uint(7-i)) != 0)
		bit := 0
		if bitSet {
			bit = 1
		}
		d.DataPin.Write(bit)
		d.ClockPin.Write(embd.High)
		d.ClockPin.Write(embd.Low)
	}

	d.DataPin.SetDirection(embd.In)

	d.ClockPin.Write(embd.High)
	ack, err := d.DataPin.Read()
	if err != nil {
		return err
	}
	if ack != embd.Low {
		return errors.New("sht1x: command not acknowledged")
	}

	d.ClockPin.Write(embd.Low)
	ack, err = d.DataPin.Read()
	if err != nil {
		return err
	}
	if ack != embd.High {
		return errors.New("sht1x: command not acknowledged")
	}

	return nil
}

func (d *SHT1x) waitForResult() error {
	d.DataPin.SetDirection(embd.In)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		ack, err := d.DataPin.Read()
		if err != nil {
			return err
		}
		if ack == embd.Low {
			return nil
		}
	}

	return errors.New("sht1x: timeout waiting for result")
}

func (d *SHT1x) getData() int {
	highByte := d.readByte()

	d.sendReadAck()

	lowByte := d.readByte()

	d.skipCrc()

	temp := highByte * 256
	temp = temp | lowByte

	return temp
}

func (d *SHT1x) readByte() int {
	d.ClockPin.SetDirection(embd.Out)
	d.DataPin.SetDirection(embd.In)

	readByte := 0
	for i := 0; i < 8; i++ {
		d.ClockPin.Write(embd.High)
		val, _ := d.DataPin.Read()
		readByte = readByte*2 + val
		d.ClockPin.Write(embd.Low)
	}
	return readByte
}

func (d *SHT1x) sendReadAck() {
	d.ClockPin.SetDirection(embd.Out)
	d.DataPin.SetDirection(embd.Out)

	d.DataPin.Write(embd.High)
	d.DataPin.Write(embd.Low)
	d.ClockPin.Write(embd.High)
	d.ClockPin.Write(embd.Low)
}

func (d *SHT1x) skipCrc() {
	d.DataPin.SetDirection(embd.Out)
	d.ClockPin.SetDirection(embd.Out)

	d.DataPin.Write(embd.High)
	d.ClockPin.Write(embd.High)
	d.ClockPin.Write(embd.Low)
}

func (d *SHT1x) rawTemperature() (float64, error) {
	d.m.Lock()
	defer d.m.Unlock()

	err := d.sendCommand(measureTemperatureCommand)
	if err != nil {
		return 0, err
	}

	err = d.waitForResult()
	if err != nil {
		return 0, err
	}

	data := d.getData()

	return float64(data), nil
}

func (d *SHT1x) rawHumidity() (float64, error) {
	d.m.Lock()
	defer d.m.Unlock()

	err := d.sendCommand(measureHumidityCommand)
	if err != nil {
		return 0, err
	}

	err = d.waitForResult()
	if err != nil {
		return 0, err
	}

	data := d.getData()

	return float64(data), nil
}

func temperatureFromRaw(rawTemperature float64) float64 {
	return d1 + d2*rawTemperature
}

func humidityFromRaw(rawHumidity, temperature float64) float64 {
	linearHumidity := c1 + c2*rawHumidity + c3*rawHumidity*rawHumidity
	relativeHumidity := (temperature-25)*(t1+t2*rawHumidity) + linearHumidity
	return relativeHumidity
}

func calculateDewPoint(humidity, temperature float64) float64 {
	tn := tnB0
	m := mB0

	if temperature > 0 {
		tn = tnA0
		m = mA0
	}

	return tn * (math.Log(humidity/100.0) + (m*temperature)/(tn+temperature)) / (m - math.Log(humidity/100.0) - m*temperature/(tn+temperature))
}
