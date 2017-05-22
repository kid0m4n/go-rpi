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

// Generic GPIO driver.

package embd

import (
	"errors"
	"fmt"
)

type pin interface {
	Close() error
}

type digitalPinFactory func(pd *PinDesc, drv GPIODriver) DigitalPin
type analogPinFactory func(pd *PinDesc, drv GPIODriver) AnalogPin
type pwmPinFactory func(pd *PinDesc, drv GPIODriver) PWMPin

type gpioDriver struct {
	pinMap PinMap

	dpf digitalPinFactory
	apf analogPinFactory
	ppf pwmPinFactory

	initializedPins map[string]pin
}

// NewGPIODriver returns a GPIODriver interface which allows control
// over the GPIO subsystem.
func NewGPIODriver(pinMap PinMap, dpf digitalPinFactory, apf analogPinFactory, ppf pwmPinFactory) GPIODriver {
	return &gpioDriver{
		pinMap: pinMap,
		dpf:    dpf,
		apf:    apf,
		ppf:    ppf,

		initializedPins: map[string]pin{},
	}
}

func (io *gpioDriver) Unregister(id string) error {
	if _, ok := io.initializedPins[id]; !ok {
		return fmt.Errorf("gpio: pin %v is not registered yet, cannot unregister", id)
	}

	delete(io.initializedPins, id)
	return nil
}

func (io *gpioDriver) DigitalPin(key interface{}) (DigitalPin, error) {
	if io.dpf == nil {
		return nil, errors.New("gpio: digital io not supported on this host")
	}

	pd, found := io.pinMap.Lookup(key, CapDigital)
	if !found {
		return nil, fmt.Errorf("gpio: could not find pin matching %v", key)
	}

	if p, ok := io.initializedPins[pd.ID]; ok {
		return p.(DigitalPin), nil
	}

	p := io.dpf(pd, io)
	io.initializedPins[pd.ID] = p

	return p, nil
}

func (io *gpioDriver) AnalogPin(key interface{}) (AnalogPin, error) {
	if io.apf == nil {
		return nil, errors.New("gpio: analog io not supported on this host")
	}

	pd, found := io.pinMap.Lookup(key, CapAnalog)
	if !found {
		return nil, fmt.Errorf("gpio: could not find pin matching %v", key)
	}

	if p, ok := io.initializedPins[pd.ID]; ok {
		return p.(AnalogPin), nil
	}

	p := io.apf(pd, io)
	io.initializedPins[pd.ID] = p

	return p, nil
}

func (io *gpioDriver) PWMPin(key interface{}) (PWMPin, error) {
	if io.ppf == nil {
		return nil, errors.New("gpio: pwm not supported on this host")
	}

	pd, found := io.pinMap.Lookup(key, CapPWM)
	if !found {
		return nil, fmt.Errorf("gpio: could not find pin matching %v", key)
	}

	if p, ok := io.initializedPins[pd.ID]; ok {
		return p.(PWMPin), nil
	}

	p := io.ppf(pd, io)
	io.initializedPins[pd.ID] = p

	return p, nil
}

func (io *gpioDriver) PinMap() PinMap {
	return io.pinMap
}

func (io *gpioDriver) Close() error {
	for _, p := range io.initializedPins {
		if err := p.Close(); err != nil {
			return err
		}
	}

	return nil
}
