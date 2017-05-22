/*
 * Copyright (c) Karan Misra 2014
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
 * associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute,
 * sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or
 * substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
 * NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// I2C support.

package embd

// I2CBus interface is used to interact with the I2C bus.
type I2CBus interface {
	// ReadByte reads a byte from the given address.
	ReadByte(addr byte) (value byte, err error)
	// ReadBytes reads a slice of bytes from the given address.
	ReadBytes(addr byte, num int) (value []byte, err error)
	// WriteByte writes a byte to the given address.
	WriteByte(addr, value byte) error
	// WriteBytes writes a slice bytes to the given address.
	WriteBytes(addr byte, value []byte) error

	// ReadFromReg reads n (len(value)) bytes from the given address and register.
	ReadFromReg(addr, reg byte, value []byte) error
	// ReadByteFromReg reads a byte from the given address and register.
	ReadByteFromReg(addr, reg byte) (value byte, err error)
	// ReadU16FromReg reads a unsigned 16 bit integer from the given address and register.
	ReadWordFromReg(addr, reg byte) (value uint16, err error)

	// WriteToReg writes len(value) bytes to the given address and register.
	WriteToReg(addr, reg byte, value []byte) error
	// WriteByteToReg writes a byte to the given address and register.
	WriteByteToReg(addr, reg, value byte) error
	// WriteU16ToReg
	WriteWordToReg(addr, reg byte, value uint16) error

	// Close releases the resources associated with the bus.
	Close() error
}

// I2CDriver interface interacts with the host descriptors to allow us
// control of I2C communication.
type I2CDriver interface {
	Bus(l byte) I2CBus

	// Close releases the resources associated with the driver.
	Close() error
}

var i2cDriverInitialized bool
var i2cDriverInstance I2CDriver

// InitI2C initializes the I2C driver.
func InitI2C() error {
	if i2cDriverInitialized {
		return nil
	}

	desc, err := DescribeHost()
	if err != nil {
		return err
	}

	if desc.I2CDriver == nil {
		return ErrFeatureNotSupported
	}

	i2cDriverInstance = desc.I2CDriver()
	i2cDriverInitialized = true

	return nil
}

// CloseI2C releases resources associated with the I2C driver.
func CloseI2C() error {
	return i2cDriverInstance.Close()
}

// NewI2CBus returns a I2CBus.
func NewI2CBus(l byte) I2CBus {
	if err := InitI2C(); err != nil {
		panic(err)
	}

	return i2cDriverInstance.Bus(l)
}
