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

func TestKernelVersionParse(t *testing.T) {
	var tests = []struct {
		versionStr          string
		major, minor, patch int
	}{
		{
			"3.8.2",
			3, 8, 2,
		},
		{
			"3.8.10+",
			3, 8, 10,
		},
	}
	for _, test := range tests {
		major, minor, patch, err := parseVersion(test.versionStr)
		if err != nil {
			t.Errorf("Failed parsing %q: %v", test.versionStr, err)
			continue
		}
		if major != test.major || minor != test.minor || patch != test.patch {
			t.Errorf("Parse of %q: got (%v, %v, %v) want (%v, %v, %v)", test.versionStr, major, minor, patch, test.major, test.minor, test.patch)
		}
	}
}
