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

// Generic LED driver.

package embd

import (
	"errors"
	"fmt"
	"strconv"
)

// LEDMap type represents a LED mapping for a host.
type LEDMap map[string][]string

type ledFactory func(string) LED

type ledDriver struct {
	ledMap LEDMap

	lf ledFactory

	initializedLEDs map[string]LED
}

// NewLEDDriver returns a LEDDriver interface which allows control
// over the LED subsystem.
func NewLEDDriver(ledMap LEDMap, lf ledFactory) LEDDriver {
	return &ledDriver{
		ledMap: ledMap,
		lf:     lf,

		initializedLEDs: map[string]LED{},
	}
}

func (d *ledDriver) lookup(k interface{}) (string, error) {
	var ks string
	switch key := k.(type) {
	case int:
		ks = strconv.Itoa(key)
	case string:
		ks = key
	case fmt.Stringer:
		ks = key.String()
	default:
		return "", errors.New("led: invalid key type")
	}

	for id := range d.ledMap {
		for _, alias := range d.ledMap[id] {
			if alias == ks {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("led: no match found for %q", k)
}

func (d *ledDriver) LED(k interface{}) (LED, error) {
	id, err := d.lookup(k)
	if err != nil {
		return nil, err
	}

	led := d.lf(id)
	d.initializedLEDs[id] = led

	return led, nil
}

func (d *ledDriver) Close() error {
	for _, led := range d.initializedLEDs {
		if err := led.Close(); err != nil {
			return err
		}
	}

	return nil
}
