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

// Generic I²C driver.

package embd

import "sync"

type i2cBusFactory func(byte) I2CBus

type i2cDriver struct {
	busMap     map[byte]I2CBus
	busMapLock sync.Mutex

	ibf i2cBusFactory
}

// NewI2CDriver returns a I2CDriver interface which allows control
// over the I²C subsystem.
func NewI2CDriver(ibf i2cBusFactory) I2CDriver {
	return &i2cDriver{
		busMap: make(map[byte]I2CBus),
		ibf:    ibf,
	}
}

func (i *i2cDriver) Bus(l byte) I2CBus {
	i.busMapLock.Lock()
	defer i.busMapLock.Unlock()

	if b, ok := i.busMap[l]; ok {
		return b
	}

	b := i.ibf(l)
	i.busMap[l] = b
	return b
}

func (i *i2cDriver) Close() error {
	for _, b := range i.busMap {
		b.Close()
	}

	return nil
}
