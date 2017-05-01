// Package hcsr501 allows interfacing with the HC-SR501 PIR Sensor.
package hcsr501

import (
	"errors"
	"sync"
	"time"

	"github.com/kidoman/embd"
)

// HCSR501 represents a HCSR501 ultrasonic range finder.
type HCSR501 struct {
	TriggerPin  embd.DigitalPin
	initialized bool
	ready       bool // Allows the sensor time to settle in.
	mu          sync.RWMutex
}

// New creates a new HCSR501 interface.
func New(triggerPin embd.DigitalPin) *HCSR501 {
	return &HCSR501{TriggerPin: triggerPin}
}

// setup initializes the GPIO and sensor.
func (d *HCSR501) setup() error {
	d.mu.RLock()
	if d.initialized {
		d.mu.RUnlock()
		return nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.TriggerPin.SetDirection(embd.In); err != nil {
		return err
	}

	// Wait 60 sec for sensor to settle down.
	time.AfterFunc(60*time.Second, func() {
		d.ready = true
	})
	d.initialized = true

	return nil
}

// Detect returns true if motion was detected.
func (d *HCSR501) Detect() (bool, error) {
	if err := d.setup(); err != nil {
		return false, err
	}

	if !d.ready {
		return false, errors.New("Sensor not ready")
	}

	// Check 3 times to be sure.
	for i := 0; i < 3; i++ {
		v, err := d.TriggerPin.Read()
		if err != nil {
			return false, err
		}
		if v == embd.Low {
			return false, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return true, nil
}

// Close.
func (d *HCSR501) Close() error {
	return d.TriggerPin.SetDirection(embd.Out)
}
