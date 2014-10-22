// Package isl29125 allows interfacing with the ISL29125 RGB light sensor
// Datasheet: http://www.intersil.com/content/dam/Intersil/documents/isl2/isl29125.pdf
package isl29125

/*

   TODO:
       - add support for the following:
           - config register 2 and 3
           - lux calculation
           - powerdown / power up
           - interupt config / monitor
*/

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

const (
	pollDelay = 250
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
	IRAdjustLow  = 0x00
	IRAdjustMed  = 0x20
	IRAdjustHigh = 0x3f
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
	RegisterDeviceID  = 0x00
	RegisterConfig1   = 0x01
	RegisterConfig2   = 0x02
	RegisterConfig3   = 0x03
	RegisterFlags     = 0x08
	RegisterGreenLow  = 0x09
	RegisterGreenHigh = 0x0a
	RegisterRedLow    = 0x0b
	RegisterRedHigh   = 0x0c
	RegisterBlueLow   = 0x0d
	RegisterBlueHigh  = 0x0e
)

const (
	FlagReady               = 0x00
	FlagInterrupt           = 0x01
	FlagConversion          = 0x02
	FlagPowerDownOrBrownOut = 0x04
	FlagGreenConverting     = 0x10
	FlagRedConverting       = 0x20
	FlagBlueConverting      = 0x30
)

const (
	DefaultConfig = ModeRGB | LuxRange375 | Resolution16Bit | SyncStartOnWrite
)

const (
	SensorAddr = 0x44
	DeviceID   = 0x7d
)

const (
	CmdGetStatus = 0x08
	CmdReset     = 0x46
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

	mode uint8
}

// New returns an ISL29125 for a given config
func New(config uint8, bus embd.I2CBus) *ISL29125 {
	glog.Info("Creating new ISL29125")
	return &ISL29125{Bus: bus, Poll: pollDelay, mode: config}
}

func (i *ISL29125) Init() error {

	// verify that i2c device is reachable on specified bus and that it reports back the correct ID
	id, err := i.Bus.ReadByteFromReg(SensorAddr, RegisterDeviceID)
	if err != nil {
		return err
	}
	if DeviceID != id {
		return fmt.Errorf("Invalid device id. Expected [%x] but device reports [%x]", DeviceID, id)
	}

	// power down device ( don't know current state; assume it's running )
	err = i.Bus.WriteByteToReg(SensorAddr, RegisterConfig1, ModePowerDown)
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	// reset device
	err = i.Bus.WriteByteToReg(SensorAddr, DeviceID, CmdReset)
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	// verify status after reset is ready
	status, err := i.Bus.ReadByteFromReg(SensorAddr, CmdGetStatus)
	if err != nil {
		return err
	}
	if status != FlagReady {
		return fmt.Errorf("Invalid device status. Expected [%x] but device reports [%x]", FlagReady, status)
	}

	// set config 1 to user specified mode
	err = i.Bus.WriteByteToReg(SensorAddr, RegisterConfig1, i.mode)
	if err != nil {
		return err
	}
	// set config 2 to fixed value
	err = i.Bus.WriteByteToReg(SensorAddr, RegisterConfig2, IRAdjustHigh)
	if err != nil {
		return err
	}

	// set config 3 to fixed value
	err = i.Bus.WriteByteToReg(SensorAddr, RegisterConfig3, 0x0)
	if err != nil {
		return err
	}

	return nil

}

func (i *ISL29125) getReading() (*Reading, error) {

	glog.Info("Getting reading")
	red, err := i.Bus.ReadWordFromReg(SensorAddr, RegisterRedLow)
	if err != nil {
		return nil, err
	}
	green, err := i.Bus.ReadWordFromReg(SensorAddr, RegisterGreenLow)
	if err != nil {
		return nil, err
	}
	blue, err := i.Bus.ReadWordFromReg(SensorAddr, RegisterBlueLow)
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
