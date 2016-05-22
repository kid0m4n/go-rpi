// Package ina219 allows interfacing with Texas Instruments INA219 current
// monitor. This is a high side current and voltage monitor with and I2C
// interfcace.

// TODO - add options to config voltage range, sensitivity and averaging
// TODO - Current and Power can overflow within normal ranges
// TODO - add documentation

package ina219

import (
	"math"
	"sync"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

const (
	configReg  = 0x00
	shuntVReg  = 0x01
	busVReg    = 0x02
	powerReg   = 0x03
	currentReg = 0x04
	calibReg   = 0x05
)

// INA219 represents an INA219 current sensor.
type INA219 struct {
	Bus embd.I2CBus

	address        byte
	shuntResitance float64

	initialized bool
	mu          sync.RWMutex
}

func New(bus embd.I2CBus, addr byte, shunt float64) *INA219 {
	ina := INA219{
		Bus:            bus,
		address:        addr,
		shuntResitance: shunt,
	}
	return &ina
}

func (d *INA219) setup() error {
	d.mu.RLock()
	if d.initialized {
		d.mu.RUnlock()
		return nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()

	//config := uint16(0x219F) // 12, bit no integration, 32v bus range , 40mV shunt range
	//config := uint16(0x2777) // 12, bit 64x integration, 32v bus range , 40mV shunt range
	config := uint16(0x2F77) // 12, bit 64x integration, 32v bus range , 80mV shunt range

	err := d.Bus.WriteWordToReg(d.address, configReg, config)
	if err != nil {
		return err
	}

	calib := uint16(0.5 + 40.96/d.shuntResitance)

	err = d.Bus.WriteWordToReg(d.address, calibReg, calib)
	if err != nil {
		return err
	}

	glog.V(1).Infof("ina219: initaliized")

	d.initialized = true

	return nil
}

func (d *INA219) ShuntVoltage() (float64, error) {
	if err := d.setup(); err != nil {
		return math.NaN(), err
	}

	v, err := d.Bus.ReadWordFromReg(d.address, shuntVReg)
	if err != nil {
		return math.NaN(), err
	}

	voltage := float64(int16(v)) / 100000.0

	return voltage, nil
}

func (d *INA219) Voltage() (float64, error) {
	if err := d.setup(); err != nil {
		return math.NaN(), err
	}

	v, err := d.Bus.ReadWordFromReg(d.address, busVReg)
	if err != nil {
		return math.NaN(), err
	}

	voltage := float64(v>>3) / 250.0

	return voltage, nil
}

func (d *INA219) Current() (float64, error) {
	if err := d.setup(); err != nil {
		return math.NaN(), err
	}

	v, err := d.Bus.ReadWordFromReg(d.address, currentReg)
	if err != nil {
		return math.NaN(), err
	}

	current := float64(int16(v)) / 1000.0

	return current, nil
}

func (d *INA219) Power() (float64, error) {
	if err := d.setup(); err != nil {
		return math.NaN(), err
	}

	v, err := d.Bus.ReadWordFromReg(d.address, powerReg)
	if err != nil {
		return math.NaN(), err
	}

	current := float64(int16(v)) / 50.0

	return current, nil
}

// Close
func (d *INA219) Close() error {
	// put in power down mode
	config := uint16(0x0000)
	err := d.Bus.WriteWordToReg(d.address, configReg, config)
	if err != nil {
		return err
	}
	return nil
}
