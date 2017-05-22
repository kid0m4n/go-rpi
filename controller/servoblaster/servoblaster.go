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

// Package servoblaster allows interfacing with the software servoblaster driver.
//
// More details on ServoBlaster at: https://github.com/richardghirst/PiBits/tree/master/ServoBlaster
package servoblaster

import (
	"fmt"
	"os"

	"github.com/golang/glog"
)

// ServoBlaster represents a software RPi PWM/PCM based servo control module.
type ServoBlaster struct {
	initialized bool
	fd          *os.File
}

// New creates a new ServoBlaster instance.
func New() *ServoBlaster {
	return &ServoBlaster{}
}

func (d *ServoBlaster) setup() error {
	if d.initialized {
		return nil
	}
	var err error
	if d.fd, err = os.OpenFile("/dev/servoblaster", os.O_WRONLY, os.ModeExclusive); err != nil {
		return err
	}
	d.initialized = true
	return nil
}

type pwmChannel struct {
	d *ServoBlaster

	channel int
}

func (p *pwmChannel) SetMicroseconds(us int) error {
	return p.d.setMicroseconds(p.channel, us)
}

func (d *ServoBlaster) Channel(channel int) *pwmChannel {
	return &pwmChannel{d: d, channel: channel}
}

// SetMicroseconds sends a command to the PWM driver to generate a us wide pulse.
func (d *ServoBlaster) setMicroseconds(channel, us int) error {
	if err := d.setup(); err != nil {
		return err
	}
	cmd := fmt.Sprintf("%v=%vus\n", channel, us)
	glog.V(1).Infof("servoblaster: sending command %q", cmd)
	_, err := d.fd.WriteString(cmd)
	return err
}

// Close closes the open driver handle.
func (d *ServoBlaster) Close() error {
	if d.fd != nil {
		return d.fd.Close()
	}
	return nil
}
