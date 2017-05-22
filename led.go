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

// LED support.

package embd

// The LED interface is used to control a led on the prototyping board.
type LED interface {
	// On switches the LED on.
	On() error

	// Off switches the LED off.
	Off() error

	// Toggle toggles the LED.
	Toggle() error

	// Close releases resources associated with the LED.
	Close() error
}

// LEDDriver interface interacts with the host descriptors to allow us
// control of the LEDs.
type LEDDriver interface {
	LED(key interface{}) (LED, error)

	Close() error
}

var ledDriverInitialized bool
var ledDriverInstance LEDDriver

// InitLED initializes the LED driver.
func InitLED() error {
	if ledDriverInitialized {
		return nil
	}

	desc, err := DescribeHost()
	if err != nil {
		return err
	}

	if desc.LEDDriver == nil {
		return ErrFeatureNotSupported
	}

	ledDriverInstance = desc.LEDDriver()
	ledDriverInitialized = true

	return nil
}

// CloseLED releases resources associated with the LED driver.
func CloseLED() error {
	return ledDriverInstance.Close()
}

// NewLED returns a LED interface which allows control over the LED.
func NewLED(key interface{}) (LED, error) {
	if err := InitLED(); err != nil {
		return nil, err
	}

	return ledDriverInstance.LED(key)
}

// LEDOn switches the LED on.
func LEDOn(key interface{}) error {
	led, err := NewLED(key)
	if err != nil {
		return err
	}

	return led.On()
}

// LEDOff switches the LED off.
func LEDOff(key interface{}) error {
	led, err := NewLED(key)
	if err != nil {
		return err
	}

	return led.Off()
}

// LEDToggle toggles the LED.
func LEDToggle(key interface{}) error {
	led, err := NewLED(key)
	if err != nil {
		return err
	}

	return led.Toggle()
}
