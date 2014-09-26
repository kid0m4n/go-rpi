// Package isl29125 allows interfacing with the ISL29125 RGB light sensor
// Datasheet: http://www.intersil.com/content/dam/Intersil/documents/isl2/isl29125.pdf
package isl29125

/*

   TODO:
       - improve constructor
       - add support for config register 2 and 3
       - add support for lux calculation
*/

import (
	"time"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

const (
	sensorI2cAddr = 0x44

	pollDelay = 250

	deviceRegisterAddr = 0x00
	conf1RegisterAddr  = 0x01
	flagRegisterAdr    = 0x08
)

const (
	ModePowerDown = 0x00
	ModeGreenOnly = 0x01
	ModeRedOnly   = 0x02
	ModeBlueOnly  = 0x03
	ModeStandBy   = 0x04
	ModeRGB       = 0x05
	ModeRG        = 0x06
	ModeGB        = 0x07
)

const (
	LuxRange375 = 0x00
	LuxRange10k = 0x08
)

const (
	Resolution16Bit = 0x00
	Resolution12Bit = 0x10
)

const (
	SyncStartOnWrite = 0x00
	SyncStartOnInt   = 0x20
)

const (
	RegisterGreenLow  = 0x09
	RegisterGreenHigh = 0x0a
	RegisterRedLow    = 0x0b
	RegisterRedHigh   = 0x0c
	RegisterBlueLow   = 0x0d
	RegisterBlueHigh  = 0x0e
)

const (
	FlagInterrupt           = 0x01
	FlagConversion          = 0x02
	FlagPowerDownOrBrownOut = 0x04
	FlagGreenConverting     = 0x10
	FlagRedConverting       = 0x20
	FlagBlueConverting      = 0x30
)

const (
	DefaultConfig = ModePowerDown | LuxRange375 | Resolution16Bit | SyncStartOnWrite
)

const (
	DeviceID = 0x7d
)

// Reading represents a single reading from an RGB light sensor
type Reading struct {
	Red        uint16
	Green      uint16
	Blue       uint16
	Lux        uint16
	LuxRange   int
	Resolution int
}

// ISL29125 represents an RGB light sensor
type ISL29125 struct {
	Bus  embd.I2CBus
	Poll int

	readings chan *Reading
	quit     chan bool

	mode       int
	luxRange   int
	resolution int
	syncStart  int
}

// New returns an ISL29125 for a given config
func New(config int, bus embd.I2CBus) *ISL29125 {
	glog.Info("Creating new ISL29125")
	return &ISL29125{Bus: bus, Poll: pollDelay}
}

func (i *ISL29125) getReading() (*Reading, error) {

	glog.Info("Getting reading")
	red, err := i.Bus.ReadWordFromReg(sensorI2cAddr, RegisterRedLow)
	if err != nil {
		return nil, err
	}
	green, err := i.Bus.ReadWordFromReg(sensorI2cAddr, RegisterGreenLow)
	if err != nil {
		return nil, err
	}
	blue, err := i.Bus.ReadWordFromReg(sensorI2cAddr, RegisterBlueLow)
	if err != nil {
		return nil, err
	}

	return &Reading{Red: red, Green: green, Blue: blue}, nil
}

// Reading returns a single sensor reading
func (i *ISL29125) Reading() (*Reading, error) {
	select {
	case r := <-i.readings:
		return r, nil
	default:
		return i.getReading()
	}
}

// Run starts continuous sensor data acquisition loop.
func (i *ISL29125) Run() {
	glog.Info("Running sensor")
	go func() {
		i.quit = make(chan bool)
		i.readings = make(chan *Reading)
		timer := time.Tick(time.Duration(i.Poll) * time.Millisecond)

		var reading *Reading

		for {
			select {
			case i.readings <- reading:
			case <-timer:
				r, err := i.getReading()
				if err == nil {
					reading = r
				}
			case <-i.quit:
				i.readings = nil
				return
			}
		}
	}()
	return
}

// Close down sensor
func (i *ISL29125) Close() {
	glog.Info("Closing sensor")
	if i.quit != nil {
		i.quit <- true
	}
	return
}
