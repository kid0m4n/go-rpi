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

package bbb

import (
	"testing"

	"github.com/cfreeman/embd"
)

func TestAnalogPinClose(t *testing.T) {
	pinMap := embd.PinMap{
		&embd.PinDesc{ID: "P1_1", Aliases: []string{"1"}, Caps: embd.CapAnalog},
	}
	driver := embd.NewGPIODriver(pinMap, nil, newAnalogPin, nil)
	pin, err := driver.AnalogPin(1)
	if err != nil {
		t.Fatalf("Looking up analog pin 1: got %v", err)
	}
	pin.Close()
	pin2, err := driver.AnalogPin(1)
	if err != nil {
		t.Fatalf("Looking up analog pin 1: got %v", err)
	}
	if pin == pin2 {
		t.Fatal("Looking up closed analog pin 1: but got the old instance")
	}
}
