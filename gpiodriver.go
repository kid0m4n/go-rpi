// Generic GPIO driver.

package embd

import (
	"errors"
	"fmt"
	"syscall"
)

const (
	MaxGPIOInterrupt = 64
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

	watchEventCallbacks map[int]InterruptPin
	initializedPins     map[string]pin
}

// NewGPIODriver returns a GPIODriver interface which allows control
// over the GPIO subsystem.
func NewGPIODriver(pinMap PinMap, dpf digitalPinFactory, apf analogPinFactory, ppf pwmPinFactory) GPIODriver {
	return &gpioDriver{
		pinMap: pinMap,
		dpf:    dpf,
		apf:    apf,
		ppf:    ppf,

		watchEventCallbacks: map[int]InterruptPin{},
		initializedPins:     map[string]pin{},
	}
}

var epollFD int

func (io *gpioDriver) initializeEpoll() {
	var err error
	epollFD, err = syscall.EpollCreate1(0)
	if err != nil {
		panic(fmt.Sprintf("Unable to create epoll FD: ", err.Error()))
	}

	go func() {
		var epollEvents [MaxGPIOInterrupt]syscall.EpollEvent

		for {
			numEvents, err := syscall.EpollWait(epollFD, epollEvents[:], -1)
			if err != nil {
				panic(fmt.Sprintf("EpollWait error: %s", err.Error()))
			}
			for i := 0; i < numEvents; i++ {
				if eventPin, exists := io.watchEventCallbacks[int(epollEvents[i].Fd)]; exists {
					eventPin.Signal()
				}
			}
		}
	}()
}

func (io *gpioDriver) RegisterInterrupt(p InterruptPin) error {

	if epollFD == 0 {
		io.initializeEpoll()
	}

	fd := p.Fd()

	var event syscall.EpollEvent
	event.Events = syscall.EPOLLIN | (syscall.EPOLLET & 0xffffffff) | syscall.EPOLLPRI

	io.watchEventCallbacks[fd] = p

	if err := syscall.SetNonblock(fd, true); err != nil {
		return err
	}

	event.Fd = int32(fd)

	if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &event); err != nil {
		return err
	}

	return nil
}

func (io *gpioDriver) UnregisterInterrupt(p InterruptPin) error {

	fd := p.Fd()

	// check if we are watching this pin
	if _, ok := io.watchEventCallbacks[fd]; !ok {
		return nil
	}

	if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
		return err
	}

	if err := syscall.SetNonblock(fd, false); err != nil {
		return err
	}

	delete(io.watchEventCallbacks, fd)

	return nil
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

func (io *gpioDriver) Close() error {
	for _, p := range io.initializedPins {
		if err := p.Close(); err != nil {
			return err
		}
	}

	return nil
}
