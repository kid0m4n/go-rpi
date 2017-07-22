// Package hmc5883l allows interfacing with the HMC5883L magnetometer.
package hmc5883l

import (
	"math"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

const (
	magAddress = 0x1E

	// Register addresses.
	magConfigRegA = 0x00
	magConfigRegB = 0x01
	magModeReg    = 0x02
	magMSBx       = 0x03
	magLSBx       = 0x04
	magMSBz       = 0x05
	magLSBz       = 0x06
	magMSBy       = 0x07
	magLSBy       = 0x08
	magStatusReg  = 0x09

	// ConfigA Params.
	MagHz75         = 0x00 // ODR = 0.75 Hz
	Mag1Hz5         = 0x04 // ODR = 1.5 Hz
	Mag3Hz          = 0x08 // ODR = 3 Hz
	Mag7Hz5         = 0x0C // ODR = 7.5 Hz
	Mag15Hz         = 0x10 // ODR = 15 Hz
	Mag30Hz         = 0x14 // ODR = 30 Hz
	Mag75Hz         = 0x18 // ODR = 75 Hz
	MagNormal       = 0x00 // Normal mode
	MagPositiveBias = 0x01 // Positive bias mode
	MagNegativeBias = 0x02 // Negative bias mode

	MagCRADefault = Mag15Hz | MagNormal // 15 Hz and normal mode is the default

	// ConfigB Params.
	MagCRBDefault = 0x20 // Gain 1090

	// Mode Reg Params.
	MagContinuous = 0x00 // Continuous conversion mode
	MagSingle     = 0x01 // Single conversion mode
	MagSleep      = 0x03 // Sleep mode

	MagMRDefault = MagContinuous // Continuous conversion is the default

	pollDelay = 250 // Delay before reading from mag. (ms)
)

type calib struct {
	minX int16
	maxX int16
	minY int16
	maxY int16
}

// HMC5883L represents a HMC5883L magnetometer.
type HMC5883L struct {
	Bus         embd.I2CBus
	Poll        int
	initialized bool
	mu          sync.RWMutex
	headings    chan float64
	quit        chan struct{}
	calibData   calib
}

// New creates a new HMC5883L interface. The bus variable controls
// the I2C bus used to communicate with the device.
func New(bus embd.I2CBus) *HMC5883L {
	return &HMC5883L{Bus: bus, Poll: pollDelay}
}

// Initialize the device
func (d *HMC5883L) setup() error {
	d.mu.RLock()
	if d.initialized {
		d.mu.RUnlock()
		return nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.Bus.WriteByteToReg(magAddress, magConfigRegA, MagCRADefault); err != nil {
		return err
	}
	if err := d.Bus.WriteByteToReg(magAddress, magConfigRegB, MagCRBDefault); err != nil {
		return err
	}
	if err := d.Bus.WriteByteToReg(magAddress, magModeReg, MagMRDefault); err != nil {
		return err
	}

	d.initialized = true

	return nil
}

func (d *HMC5883L) measureHeading() (float64, error) {
	if err := d.setup(); err != nil {
		return 0, err
	}

	data := make([]byte, 6)
	if err := d.Bus.ReadFromReg(magAddress, magMSBx, data); err != nil {
		return 0, err
	}

	x := int16(data[0])<<8 | int16(data[1])
	z := int16(data[2])<<8 | int16(data[3])
	y := int16(data[4])<<8 | int16(data[5])

	/*Note on Calibration:
	   In order to compensate for tilt of compass, it has to be calibrated. To calibrate
		 rotate the compass a full 360'. Then calculate the X and Y offsets as
		 Xoffset = (minX + maxX)/2 ; Yoffset = (minY +maxY)/2
		 when reading the raw values update them by offset
		 Xadj = Xraw - Xoffset
		 Yadj = Yraw - Yoffset
	*/
	if x < d.calibData.minX {
		d.calibData.minX = x
	}

	if x > d.calibData.maxX {
		d.calibData.maxX = x
	}

	if y < d.calibData.minY {
		d.calibData.minY = y
	}

	if y > d.calibData.maxY {
		d.calibData.maxY = y
	}

	x -= 274
	y -= 56

	heading := math.Atan2(float64(y), float64(x))
	heading += 233.9 / 1000
	if heading < 0 {
		heading += 2 * math.Pi
	}

	if heading > 2*math.Pi {
		heading -= 2 * math.Pi
	}

	head := heading * 180 / math.Pi

	glog.V(3).Infof("Mag X=%v Y=%v Z=%v HEAD=%v CalibData=%v", x, y, z, head, d.calibData)
	return head, nil
}

// Heading returns the current heading [0, 360).
func (d *HMC5883L) Heading() (float64, error) {
	select {
	case heading := <-d.headings:
		return heading, nil
	default:
		glog.V(3).Infof("lsm303: no headings available... measuring")
		return d.measureHeading()
	}
}

// Run starts the sensor data acquisition loop.
func (d *HMC5883L) Run() error {
	go func() {
		d.quit = make(chan struct{})

		timer := time.Tick(time.Duration(d.Poll) * time.Millisecond)

		var heading float64

		for {
			select {
			case <-timer:
				h, err := d.measureHeading()
				if err == nil {
					heading = h
				}
				if err == nil && d.headings == nil {
					d.headings = make(chan float64)
				}
			case d.headings <- heading:
			case <-d.quit:
				d.headings = nil
				return
			}

		}
	}()

	return nil
}

// Close the sensor data acquisition loop and put the HMC5883L into sleep mode.
func (d *HMC5883L) Close() error {
	if d.quit != nil {
		d.quit <- struct{}{}
	}
	return d.Bus.WriteByteToReg(magAddress, magModeReg, MagSleep)
}
