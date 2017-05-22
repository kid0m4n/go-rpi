/*
 * Copyright (c) Kunal Powar 2014
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

package embd

import "sync"

type spiBusFactory func(int, byte, byte, int, int, int, func() error) SPIBus

type spiDriver struct {
	spiDevMinor int
	initializer func() error

	busMap     map[byte]SPIBus
	busMapLock sync.Mutex

	sbf spiBusFactory
}

// NewSPIDriver returns a SPIDriver interface which allows control
// over the SPI bus.
func NewSPIDriver(spiDevMinor int, sbf spiBusFactory, i func() error) SPIDriver {
	return &spiDriver{
		spiDevMinor: spiDevMinor,
		sbf:         sbf,
		initializer: i,
	}
}

// Bus returns a SPIBus interface which allows us to use spi functionalities
func (s *spiDriver) Bus(mode, channel byte, speed, bpw, delay int) SPIBus {
	s.busMapLock.Lock()
	defer s.busMapLock.Unlock()

	b := s.sbf(s.spiDevMinor, mode, channel, speed, bpw, delay, s.initializer)
	s.busMap = make(map[byte]SPIBus)
	s.busMap[channel] = b
	return b
}

// Close cleans up all the initialized SPIbus
func (s *spiDriver) Close() error {
	for _, b := range s.busMap {
		b.Close()
	}

	return nil
}
