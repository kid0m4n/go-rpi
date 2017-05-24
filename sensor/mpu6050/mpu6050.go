/*
Package mpu6050 allows interacting with mpu6050 gyroscoping/acceleration sensor.
The interface is as close as possible to the sensor's firmware datasheet with no extra magic.
https://www.invensense.com/wp-content/uploads/2015/02/MPU-6000-Register-Map1.pdf
*/
package mpu6050

import (
	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

// Direct constants mapping with the register names
const (
	I2C_ADDR   = 0x68
	PWR_MGMT_1 = 0x6b

	// CONFIG is used to set the digital low pass filter (DLPF)
	CONFIG       = 0x1A
	GYRO_CONFIG  = 0x1B
	ACCEL_CONFIG = 0x1C

	DATA = 0x3B
	// ACCEL_XOUT_L = 0x3C
	// ACCEL_YOUT_H = 0x3D
	// ACCEL_YOUT_L = 0x3E
	// ACCEL_ZOUT_H = 0x3F
	// ACCEL_ZOUT_L = 0x40
	// GYRO_XOUT_H = 0x43
	// GYRO_XOUT_L = 0x44
	// GYRO_YOUT_H = 0x45
	// GYRO_YOUT_L = 0x46
	// GYRO_ZOUT_H = 0x47
	// GYRO_ZOUT_L = 0x48
	// TEMP_OUT_H  = 0x41
	// TEMP_OUT_L  = 0x42
	//Temperature in degrees C = (TEMP_OUT Register Value as a signed quantity)/340 + 36.53

	SIGNAL_PATH_RESET = 0x68

	WAKE         = 0x0
	SLEEP        = 0x40
	DEVICE_RESET = 0x80
)

type reading struct {
	highReg, lowReg uint8
	devider         float64
}

func (r *reading) decimal() float64 {
	d := float64(uint16(r.highReg)<<8 | uint16(r.lowReg))
	if d >= 0x8000 {
		d = -((65535 - d) + 1)
	}
	d /= r.devider
	return d

}

type orientation struct {
	X, Y, Z float64
}
type velocity struct {
	X, Y, Z float64
}

type temperature struct {
	Celsius float64
}

type rotation struct {
	X, Y float64
}

type MPU6050Reading struct {
	aReading orientation
	gReading velocity
	tReading temperature
}

func (r *MPU6050Reading) Orientation() orientation {
	return r.aReading
}
func (r *MPU6050Reading) Velocity() velocity {
	return r.gReading
}

func (r *MPU6050Reading) Temp() temperature {
	return r.tReading
}

// Config represents a MPU6050 range setting.
type Config struct {
	GiroScale  string
	AccelScale string
	Dlpf       string
}

// scaleRange represents a range unit where selection is the register value and sensitivity is the range devider
type scaleRange struct {
	name        string
	selection   byte
	sensitivity float64
}

// used to validate the config passed from the initialization and map it to the internal Digital Low Pass Filter (DLPF) values.
var dlpfRanges = map[string]byte{
	"0": 0x0,
	"1": 0x1,
	"2": 0x2,
	"3": 0x3,
	"4": 0x4,
	"5": 0x5,
	"6": 0x6}

// used to parse the config passed from the initialization and map it to the internal scale range values.
var gyroRanges = map[string]scaleRange{
	"250":  scaleRange{"250 °/s", 0x0, 131},
	"500":  scaleRange{"500 °/s", 0x8, 65.5},
	"1000": scaleRange{"1000 °/s", 0x10, 32.8},
	"2000": scaleRange{"2000 °/s", 0x18, 16.4},
}

// used to parse the config passed from the initialization and map it to the internal scale range values.
var accelRanges = map[string]scaleRange{
	"2g":  scaleRange{"2g", 0x0, 16384},
	"4g":  scaleRange{"4g", 0x8, 8192},
	"8g":  scaleRange{"8g", 0x10, 4096},
	"16g": scaleRange{"16g", 0x18, 2048},
}

func readingDevider(b embd.I2CBus, reg byte, s map[string]scaleRange) float64 {
	scale, _ := b.ReadByteFromReg(I2C_ADDR, reg)
	scale = (scale & 0x18) // mask the bits that are not of interest

	var devider float64 = 1
	for _, v := range s {
		if v.selection == scale {
			devider = v.sensitivity
			break
		}
	}
	return devider
}

// MPU6050 represents a MPU6050 3-axis gyroscope and acceleromoter.
type MPU6050 struct {
	Bus        embd.I2CBus
	gyroRange  scaleRange
	accelRange scaleRange
	dlpf       byte
}

// New creates a new MPU6050 interface.
func New(bus embd.I2CBus, c *Config) (*MPU6050, error) {
	var s *MPU6050
	var gR = gyroRanges["250"]
	var aR = accelRanges["2g"]
	var dlpf = dlpfRanges["0"]

	if r, ok := gyroRanges[c.GiroScale]; ok {
		gR = r
	} else if c.GiroScale != "" {
		glog.Fatalln("Incorrect gyro scale range value!")
	}
	if r, ok := accelRanges[c.AccelScale]; ok {
		aR = r
	} else if c.AccelScale != "" {
		glog.Fatalln("Incorrect accel scale range value!")
	}
	if r, ok := dlpfRanges[c.Dlpf]; ok {
		dlpf = r
	} else if c.Dlpf != "" {
		glog.Fatalln("Incorrect dlpf value!")
	}
	s = &MPU6050{
		Bus:        bus,
		gyroRange:  gR,
		accelRange: aR,
		dlpf:       dlpf,
	}

	return s, nil
}

// Wake writes to the power management register to disable the sleep mode
func (s *MPU6050) Start() error {
	if err := s.wake(); err != nil {
		return err
	}
	if err := s.setup(); err != nil {
		return err
	}
	return nil
}

// Wake writes to the power management register to disable the sleep mode
func (s *MPU6050) wake() error {
	if err := s.Bus.WriteByteToReg(I2C_ADDR, PWR_MGMT_1, WAKE); err != nil {
		return err
	}
	glog.V(1).Infof("mpu6050: sensor is now alive")
	return nil
}

// Close writes to the power management register to reset the sensor and put it in sleep mode
func (s *MPU6050) Close() error {
	if err := s.Bus.WriteByteToReg(I2C_ADDR, PWR_MGMT_1, SLEEP|DEVICE_RESET); err != nil {
		return err
	}
	glog.V(1).Infof("mpu6050: sensor is reset and put to sleep")
	return nil
}

func (s *MPU6050) setup() error {
	if err := s.Bus.WriteByteToReg(I2C_ADDR, GYRO_CONFIG, s.gyroRange.selection); err != nil {
		return err
	}
	glog.V(1).Infof("mpu6050: sensor gyro scale is %s", s.gyroRange.name)

	if err := s.Bus.WriteByteToReg(I2C_ADDR, ACCEL_CONFIG, s.accelRange.selection); err != nil {
		return err
	}
	glog.V(1).Infof("mpu6050: sensor gyro scale is %s", s.accelRange.name)

	if err := s.Bus.WriteByteToReg(I2C_ADDR, CONFIG, s.dlpf); err != nil {
		return err
	}
	glog.V(1).Infof("mpu6050: sensor digital low pass filter is %d", s.dlpf)

	return nil
}

func (s *MPU6050) Read() *MPU6050Reading {
	r := &MPU6050Reading{}
	// read all 6 registers at once to make sure we are getting the correct reading sample sequence
	data := make([]byte, 14)
	if err := s.Bus.ReadFromReg(I2C_ADDR, DATA, data); err != nil {
		glog.Fatalln(err)
	}

	// take the devider for every reading in case another app has changed the sensor sensitivity
	deviderA := readingDevider(s.Bus, ACCEL_CONFIG, accelRanges)

	xA := reading{uint8(data[0]), uint8(data[1]), deviderA}
	yA := reading{uint8(data[2]), uint8(data[3]), deviderA}
	zA := reading{uint8(data[4]), uint8(data[5]), deviderA}

	r.aReading = orientation{xA.decimal(), yA.decimal(), zA.decimal()}

	t := reading{uint8(data[6]), uint8(data[7]), 1}
	r.tReading = temperature{Celsius: (t.decimal()/340 + 36.53)}
	deviderG := readingDevider(s.Bus, GYRO_CONFIG, gyroRanges)
	xG := reading{uint8(data[8]), uint8(data[9]), deviderG}
	yG := reading{uint8(data[10]), uint8(data[11]), deviderG}
	zG := reading{uint8(data[12]), uint8(data[13]), deviderG}

	r.gReading = velocity{xG.decimal(), yG.decimal(), zG.decimal()}

	return r
}
