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

package util

import "testing"

func TestMap(t *testing.T) {
	var tests = []struct {
		x, inmin, inmax, outmin, outmax int64
		val                             int64
	}{
		{
			90, 0, 180, 1000, 2000,
			1500,
		},
		{
			10, 10, 15, 10, 20,
			10,
		},
		{
			15, 10, 15, 10, 20,
			20,
		},
	}
	for _, test := range tests {
		val := Map(test.x, test.inmin, test.inmax, test.outmin, test.outmax)
		if val != test.val {
			t.Errorf("Map of %v from (%v -> %v) to (%v -> %v): got %v, want %v", test.x, test.inmin, test.inmax, test.outmin, test.outmax, val, test.val)
		}
	}
}
