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

package embd

import "testing"

func TestPinMapLookup(t *testing.T) {
	var tests = []struct {
		key interface{}
		cap int
		id  string

		found bool
	}{
		{"10", CapAnalog, "P1_1", true},
		{10, CapAnalog, "P1_1", true},
		{"10", CapDigital, "P1_2", true},
		{"P1_2", CapDigital, "P1_2", true},
		{"P1_2", CapAnalog, "P1_2", true},
		{"GPIO10", CapDigital, "P1_2", true},
		{key: "NOTTHERE", found: false},
	}
	var pinMap = PinMap{
		&PinDesc{ID: "P1_1", Aliases: []string{"AN1", "10"}, Caps: CapAnalog},
		&PinDesc{ID: "P1_2", Aliases: []string{"10", "GPIO10"}, Caps: CapDigital},
	}
	for _, test := range tests {
		pd, found := pinMap.Lookup(test.key, test.cap)
		if found != test.found {
			t.Errorf("Outcome mismatch for %v: got found = %v, expected found = %v", test.key, found, test.found)
			continue
		}
		if !found {
			continue
		}
		if pd.ID != test.id {
			var capStr string
			switch test.cap {
			case CapDigital:
				capStr = "CapDigital"
			case CapAnalog:
				capStr = "CapAnalog"
			default:
				t.Fatalf("Unknown cap %v", test.cap)
			}
			t.Errorf("Looking up %q with %v: got %v, want %v", test.key, capStr, pd.ID, test.id)
		}
	}
}

func BenchmarkPinMapLookup(b *testing.B) {
	var pinMap = PinMap{
		&PinDesc{ID: "P1_1", Aliases: []string{"AN1", "10"}, Caps: CapAnalog},
		&PinDesc{ID: "P1_2", Aliases: []string{"10", "GPIO10"}, Caps: CapDigital},
	}
	for i := 0; i < b.N; i++ {
		pinMap.Lookup("GPIO10", CapDigital)
	}
}
