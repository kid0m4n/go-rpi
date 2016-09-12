// Package mcp9808 is a driver for the MCP9808 temperature sensor
// and all code is based off of the documentation found here:
// http://ww1.microchip.com/downloads/en/DeviceDoc/25095A.pdf
package mcp9808

import (
	"log"
	"sync"
	"time"

	"github.com/kidoman/embd"
)

const (
	// default I2C address for device
	address = 0x18

	// Register addresses.
	regConfig       = 0x01
	regUpperTemp    = 0x02
	regLowerTemp    = 0x03
	regCriticalTemp = 0x04
	regAmbientTemp  = 0x05
	regManufID      = 0x06
	regDeviceID     = 0x07
	regResolution   = 0x08
)

const (
	configAlertMode uint16 = 1 << iota
	configAlertPolarity
	configAlertSelect
	configAlertControl
	configAlertStatus
	configInterruptClear
	configWindowTempLock
	configCriticalTempLock
	configShutDown
)

// MCP9808 represents a MCP9808 temperature sensor.
type MCP9808 struct {
	Bus  embd.I2CBus
	Poll time.Duration

	cmu sync.RWMutex

	config uint16
	synced bool

	temps chan uint16
}

// New returns a handle to a MCP9808 sensor.
func New(bus embd.I2CBus) *MCP9808 {
	return &MCP9808{Bus: bus, Poll: time.Second, synced: false}
}

// ManufacturerID reads the device manufacturer ID
func (d *MCP9808) ManufacturerID() (uint16, error) {
	d.cmu.Lock()
	defer d.cmu.Unlock()

	return d.Bus.ReadWordFromReg(address, regManufID)
}

// DeviceID reads the device ID and revision
func (d *MCP9808) DeviceID() (uint8, uint8, error) {
	d.cmu.Lock()
	defer d.cmu.Unlock()

	devIDRev, err := d.Bus.ReadWordFromReg(address, regDeviceID)
	if err != nil {
		return 0, 0, err
	}

	return uint8(devIDRev >> 8), uint8(devIDRev & 0xFF), nil
}

// Config returns the current word value of the sensor config struct and whether
// that value reflects what is set on the sensor
func (d *MCP9808) Config() (uint16, bool) {
	return d.config, d.synced
}

// ReadConfig reads the config word from the device and writes it to the config attribute
// this overwrites any changes that may have been made to the config attribute
func (d *MCP9808) ReadConfig() (uint16, error) {
	d.cmu.RLock()
	defer d.cmu.RUnlock()
	config, err := d.Bus.ReadWordFromReg(address, regConfig)
	if err != nil {
		return 0, err
	}
	d.config = config
	d.synced = true
	return d.config, nil
}

// WriteConfig writes the sensor's config word to the device and returns the resulting config
func (d *MCP9808) WriteConfig() (uint16, error) {
	d.cmu.Lock()
	if err := d.Bus.WriteWordToReg(address, regConfig, d.config); err != nil {
		return 0, err
	}
	d.cmu.Unlock()

	// read the config after write in case some changes were invalid
	return d.ReadConfig()
}

// flipConfig bit sets (1, set = true) or clears (0, set = false) a bit within the config word
func (d *MCP9808) flipConfigBit(val uint16, set bool) {
	if set {
		d.config |= val
	} else {
		d.config &= ^val
	}
	d.synced = false
}

func (d *MCP9808) readConfigValue(val uint16) bool {
	return !(d.config&(1<<val) == 0)
}

// Hysteresis applies for decreasing temperature only (hot to cold) or as temperature
// drifts below the specified limit.
type Hysteresis uint8

const (
	// Zero hysteresis represents 0 degrees Celsius of hysteresis compensation
	Zero Hysteresis = iota
	// Plus1pt5 hysteresis represents +1.5 degrees Celsius of hysteresis compensation
	Plus1pt5
	// Plus3pt0 hysteresis represents +3.0 degrees Celsius of hysteresis compensation
	Plus3pt0
	// Plus6pt0 hysteresis represents +6.0 degrees Celsius of hysteresis compensation
	Plus6pt0
)

// TempHysteresis - TUPPER and TLOWER Limit Hysteresis bits
// 00 = 0°C (power-up default)
// 01 = +1.5°C
// 10 = +3.0°C
// 11 = +6.0°C
// The hysteresis applies for decreasing temperature only (hot to cold) or as temperature
// drifts below the specified limit.
func (d *MCP9808) TempHysteresis() Hysteresis {
	return Hysteresis(d.config >> 9)
}

// SetTempHysteresis - TUPPER and TLOWER Limit Hysteresis bits
// 00 = 0°C (power-up default)
// 01 = +1.5°C
// 10 = +3.0°C
// 11 = +6.0°C
// The hysteresis applies for decreasing temperature only (hot to cold) or as temperature
// drifts below the specified limit.
// This bit can not be altered when either of the Lock bits are set (bit 6 and bit 7).
// Thi s bit can be programmed in Shutdown mode.
func (d *MCP9808) SetTempHysteresis(val Hysteresis) {
	d.config = d.config - d.config&^(d.config>>9) + uint16(val)<<9
	d.synced = false
}

// ShutdownMode bit
// 0 (false) = Continuous conversion (power-up default)
// 1 (true) = Shutdown (Low-Power mode)
// In shutdown, all power-consuming activities are disabled, though all registers can be written to or read.
func (d *MCP9808) ShutdownMode() bool {
	return d.readConfigValue(configShutDown)
}

// SetShutdownMode bit
// 0 (false) = Continuous conversion (power-up default)
// 1 (true) = Shutdown (Low-Power mode)
// In shutdown, all power-consuming activities are disabled, though all registers can be written to or read.
// This bit cannot be set to ‘1’ when either of the Lock bits is set (bit 6 and bit 7). However, it can be
// cleared to ‘0’ for continuous conversion while locked
func (d *MCP9808) SetShutdownMode(set bool) {
	d.flipConfigBit(configShutDown, set)
}

// CriticalTempLock - TCRIT Lock bit
// 0 (false) = Unlocked. TCRIT register can be written (power-up default)
// 1 (true) = Locked. TCRIT register can not be written
// When enabled, this bit remains set to ‘1’ or locked until cleared by an internal Reset
func (d *MCP9808) CriticalTempLock() bool {
	return d.readConfigValue(configCriticalTempLock)
}

// SetCriticalTempLock - TCRIT Lock bit
// 0 (false) = Unlocked. TCRIT register can be written (power-up default)
// 1 (true) = Locked. TCRIT register can not be written
// When enabled, this bit remains set to ‘1’ or locked until cleared by an internal Reset
// This bit can be programmed in Shutdown mode.
func (d *MCP9808) SetCriticalTempLock(locked bool) {
	d.flipConfigBit(configCriticalTempLock, locked)
}

// WindowTempLock - TUPPER and TLOWER Window Lock bit
// 0 (false) = Unlocked; TUPPER and TLOWER registers can be written (power-up default)
// 1 (true) = Locked; TUPPER and TLOWER registers can not be written
// When enabled, this bit remains set to ‘1’ or locked until cleared by a Power-on Reset
func (d *MCP9808) WindowTempLock() bool {
	return d.readConfigValue(configWindowTempLock)
}

// SetWindowTempLock - TUPPER and TLOWER Window Lock bit
// 0 (false) = Unlocked; TUPPER and TLOWER registers can be written (power-up default)
// 1 (true) = Locked; TUPPER and TLOWER registers can not be written
// When enabled, this bit remains set to ‘1’ or locked until cleared by a Power-on Reset
// This bit can be programmed in Shutdown mode.
func (d *MCP9808) SetWindowTempLock(locked bool) {
	d.flipConfigBit(configWindowTempLock, locked)
}

// InterruptClear - Interrupt Clear bit
// 0 (false) = No effect (power-up default)
// 1 (true) = Clear interrupt output; when read, this bit returns to ‘0’
func (d *MCP9808) InterruptClear() bool {
	return d.readConfigValue(configInterruptClear)
}

// SetInterruptClear - Interrupt Clear bit
// 0 (false) = No effect (power-up default)
// 1 (true) = Clear interrupt output; when read, this bit returns to ‘0’
// This bit can not be set to ‘1’ in Shutdown mode, but it can be cleared after the device enters Shutdown
// mode.
func (d *MCP9808) SetInterruptClear(set bool) {
	d.flipConfigBit(configInterruptClear, set)
}

// AlertStatus Alert Output Status bit
// 0 (false) = Alert output is not asserted by the device (power-up default)
// 1 (true) = Alert output is asserted as a comparator/Interrupt or critical temperature output
func (d *MCP9808) AlertStatus() bool {
	return d.readConfigValue(configAlertStatus)
}

// SetAlertStatus Alert Output Status bit
// 0 (false) = Alert output is not asserted by the device (power-up default)
// 1 (true) = Alert output is asserted as a comparator/Interrupt or critical temperature output
// This bit can not be set to ‘1’ or cleared to ‘0’ in Shutdown mode. However, if the Alert output is configured
// as Interrupt mode, and if the host controller clears to ‘0’, the interrupt, using bit 5 while the device
// is in Shutdown mode, then this bit will also be cleared ‘0’.
func (d *MCP9808) SetAlertStatus(set bool) {
	d.flipConfigBit(configAlertStatus, set)
}

// AlertControl - Alert Output Control bit
// 0 (false) = Disabled (power-up default)
// 1 (true) = Enabled
func (d *MCP9808) AlertControl() bool {
	return d.readConfigValue(configAlertControl)
}

// SetAlertControl - Alert Output Control bit
// 0 (false) = Disabled (power-up default)
// 1 (true) = Enabled
// This bit can not be altered when either of the Lock bits are set (bit 6 and bit 7).
// This bit can be programmed in Shutdown mode, but the Alert output will not assert or deassert.
func (d *MCP9808) SetAlertControl(set bool) {
	d.flipConfigBit(configAlertControl, set)
}

// AlertSelect - Alert Output Select bit
// 0 (false) = Alert output for TUPPER, TLOWER and TCRIT (power-up default)
// 1 (true) = TA > TCRIT only (TUPPER and TLOWER temperature boundaries are disabled)
func (d *MCP9808) AlertSelect() bool {
	return d.readConfigValue(configAlertSelect)
}

// SetAlertSelect - Alert Output Select bit
// 0 (false) = Alert output for TUPPER, TLOWER and TCRIT (power-up default)
// 1 (true) = TA > TCRIT only (TUPPER and TLOWER temperature boundaries are disabled)
// When the Alarm Window Lock bit is set, this bit cannot be altered until unlocked (bit 6).
// This bit can be programmed in Shutdown mode, but the Alert output will not assert or deassert.
func (d *MCP9808) SetAlertSelect(set bool) {
	d.flipConfigBit(configAlertSelect, set)
}

// AlertPolarity - Alert Output Polarity bit
// 0 (false) = Active-low (power-up default; pull-up resistor required)
// 1 (true) = Active-high
func (d *MCP9808) AlertPolarity() bool {
	return d.readConfigValue(configAlertPolarity)
}

// SetAlertPolarity - Alert Output Polarity bit
// 0 (false) = Active-low (power-up default; pull-up resistor required)
// 1 (true) = Active-high
// This bit cannot be altered when either of the Lock bits are set (bit 6 and bit 7).
// This bit can be programmed in Shutdown mode, but the Alert output will not assert or deassert.
func (d *MCP9808) SetAlertPolarity(set bool) {
	d.flipConfigBit(configAlertPolarity, set)
}

// AlertMode - Alert Output Mode bit
// 0 (false) = Comparator output (power-up default)
// 1 (true) = Interrupt output
func (d *MCP9808) AlertMode() bool {
	return d.readConfigValue(configAlertMode)
}

// SetAlertMode - Alert Output Mode bit
// 0 (false) = Comparator output (power-up default)
// 1 (true) = Interrupt output
// This bit cannot be altered when either of the Lock bits are set (bit 6 and bit 7).
// This bit can be programmed in Shutdown mode, but the Alert output will not assert or deassert.
func (d *MCP9808) SetAlertMode(set bool) {
	d.flipConfigBit(configAlertMode, set)
}

// Temperature contains the ambient temperature along with alert values.
type Temperature struct {
	CelsiusDeg                            float64
	AboveCritical, AboveUpper, BelowLower bool
}

// AmbientTemp reads the current sensor value along with the flags denoting what boundaries the
// current temperature exceeds.
func (d *MCP9808) AmbientTemp() (*Temperature, error) {
	temp, err := d.Bus.ReadWordFromReg(address, regAmbientTemp)
	if err != nil {
		return nil, err
	}

	tempC := temp &^ 0xF000

	wholeNum := float64(tempC&^0xF) / 16.0
	fraction := tempC &^ 0xFF0

	tempResult := &Temperature{
		AboveCritical: !(temp&(1<<15) == 0),
		AboveUpper:    !(temp&(1<<14) == 0),
		BelowLower:    !(temp&(1<<13) == 0)}
	tempResult.CelsiusDeg = wholeNum + (float64(fraction) / 10000.0)
	if !(temp&(1<<12) == 0) { // read sign bit
		tempResult.CelsiusDeg *= -1
	}
	return tempResult, nil
}

// readTempC reads from the ambient temperature register and returns the current temperature value in celsius
func (d *MCP9808) readTempC(reg byte) (float64, error) {
	temp, err := d.Bus.ReadWordFromReg(address, reg)
	if err != nil {
		return 0, err
	}

	log.Printf("\nTA vs TCrit: %v\nTA vs TUpper: %v\nTA vs TLower: %v\n", !(temp&(1<<15) == 0), !(temp&(1<<14) == 0), !(temp&(1<<13) == 0))

	wholeTempC := temp &^ 0xF000

	wholeNum := float64(wholeTempC&^0xF) / 16.0
	fraction := wholeTempC &^ 0xFF0

	tempRead := wholeNum + (float64(fraction) / 10000.0)
	if !(temp&(1<<12) == 0) { // read sign bit
		tempRead *= -1
	}
	return tempRead, nil
}

func (d *MCP9808) setTemp(reg byte, newTemp float64) error {
	var signBit uint16
	if newTemp < 0 {
		newTemp *= -1
		signBit = 0x1000
	}
	newTempWord := uint16(newTemp)*15 + uint16(newTemp*16) + signBit

	if err := d.Bus.WriteWordToReg(address, reg, newTempWord); err != nil {
		return err
	}
	return nil
}

// CriticalTempUpper reads the current temperature set in the critical temperature register.
func (d *MCP9808) CriticalTempUpper() (float64, error) {
	return d.readTempC(regCriticalTemp)
}

// SetCriticalTemp when the temperature goes above the set value the alert will be
// triggered if enabled.
func (d *MCP9808) SetCriticalTemp(newTemp float64) error {
	d.cmu.Lock()
	defer d.cmu.Unlock()

	d.SetCriticalTempLock(false)
	if _, err := d.WriteConfig(); err != nil {
		return err
	}

	if err := d.setTemp(regCriticalTemp, newTemp); err != nil {
		return err
	}

	d.SetCriticalTempLock(true)
	_, err := d.WriteConfig()
	return err
}

// WindowTempUpper reads the current temperature set in the upper window temperature register.
func (d *MCP9808) WindowTempUpper() (float64, error) {
	return d.readTempC(regUpperTemp)
}

// SetWindowTempUpper when the temperature goes above the set value the alert will be
// triggered if enabled.
func (d *MCP9808) SetWindowTempUpper(newTemp float64) error {
	d.cmu.Lock()
	defer d.cmu.Unlock()

	d.SetWindowTempLock(false)
	if _, err := d.WriteConfig(); err != nil {
		return err
	}

	if err := d.setTemp(regUpperTemp, newTemp); err != nil {
		return err
	}

	d.SetWindowTempLock(true)
	_, err := d.WriteConfig()
	return err
}

// WindowTempLower reads the current temperature set in the lower window temperature register.
func (d *MCP9808) WindowTempLower() (float64, error) {
	return d.readTempC(regLowerTemp)
}

// SetWindowTempLower when the temperature goes below the set value the alert will be
// triggered if enabled.
func (d *MCP9808) SetWindowTempLower(newTemp float64) error {
	d.cmu.Lock()
	defer d.cmu.Unlock()

	d.SetWindowTempLock(false)
	if _, err := d.WriteConfig(); err != nil {
		return err
	}

	if err := d.setTemp(regLowerTemp, newTemp); err != nil {
		return err
	}

	d.SetWindowTempLock(true)
	_, err := d.WriteConfig()
	return err
}

// TempResolution reads the current temperature accuracy from the sensor (affects temperature read speed)
// 0 - +/- .5 degrees C (~30ms)
// 1 - +/- .25 degrees C (~65ms)
// 2 - +/- .125 degrees C (~130ms)
// 3 (default) - +/- .0625 degrees C (~250ms)
type TempResolution uint8

const (
	HalfC TempResolution = iota
	QuarterC
	Eighth
	Sixteenth
)

// TempResolution reads the temperature resolution from the sensor.
func (d *MCP9808) TempResolution() (TempResolution, error) {
	d.cmu.Lock()
	defer d.cmu.Unlock()

	res, err := d.Bus.ReadByteFromReg(address, regResolution)
	return TempResolution(res), err
}

// SetTempResolution writes a new temperature resolution to the sensor
func (d *MCP9808) SetTempResolution(res TempResolution) error {
	d.cmu.Lock()
	defer d.cmu.Unlock()

	return d.Bus.WriteByteToReg(address, regResolution, byte(res))
}
