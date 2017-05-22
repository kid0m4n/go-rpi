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

// Pin mapping support.

package embd

import (
	"fmt"
	"strconv"
)

const (
	// CapDigital represents the digital IO capability.
	CapDigital int = 1 << iota

	// CapI2C represents pins with the I2C capability.
	CapI2C

	// CapUART represents pins with the UART capability.
	CapUART

	// CapSPI represents pins with the SPI capability.
	CapSPI

	// CapGPMS represents pins with the GPMC capability.
	CapGPMC

	// CapLCD represents pins used to carry LCD data.
	CapLCD

	// CapPWM represents pins with PWM capability.
	CapPWM

	// CapAnalog represents pins with analog IO capability.
	CapAnalog
)

// PinDesc represents a pin descriptor.
type PinDesc struct {
	ID      string
	Aliases []string
	Caps    int

	DigitalLogical int
	AnalogLogical  int
}

// PinMap type represents a collection of pin descriptors.
type PinMap []*PinDesc

// Lookup returns a pin descriptor matching the provided key and capability
// combination. This allows the same keys to be used across pins with differing
// capabilities. For example, it is perfectly fine to have:
//
//	pin1: {Aliases: [10, GPIO10], Cap: CapDigital}
//	pin2: {Aliases: [10, AIN0], Cap: CapAnalog}
//
// Searching for 10 with CapDigital will return pin1 and searching for
// 10 with CapAnalog will return pin2. This makes for a very pleasant to use API.
func (m PinMap) Lookup(k interface{}, cap int) (*PinDesc, bool) {
	var ks string
	switch key := k.(type) {
	case int:
		ks = strconv.Itoa(key)
	case string:
		ks = key
	case fmt.Stringer:
		ks = key.String()
	default:
		return nil, false
	}

	for i := range m {
		pd := m[i]

		if pd.ID == ks {
			return pd, true
		}

		for j := range pd.Aliases {
			if pd.Aliases[j] == ks && pd.Caps&cap != 0 {
				return pd, true
			}
		}
	}

	return nil, false
}
