// PCAL953A, low volage GPIO expander as found in the Raspberry
// Pi Relay board by Seeed Studio.
//
// http://wiki.seeedstudio.com/wiki/Raspberry_Pi_Relay_Board_v1.0
package pcal9535a

import (
	"github.com/kidoman/embd"
)

const (
	REG_MODE = 0x06
)

type PCAL9535A struct {
	Bus  embd.I2CBus
	Addr byte
	D    byte
}

// New creates and connects to a PCAL9535A GPIO expander.
func New(bus embd.I2CBus, addr byte) (*PCAL9535A, error) {
	return &PCAL9535A{
		Bus:  bus,
		Addr: addr,
		D:    0xff,
	}, bus.WriteByteToReg(addr, REG_MODE, 0xff)
}

// Sets the nominated GPIO pin to either high (on = true) or low (on = false)
func (c *PCAL9535A) SetPin(pin uint, on bool) error {
	if on {
		c.D &= ^(byte(0x1) << pin)
	} else {
		c.D |= (byte(0x1) << pin)
	}

	return c.Bus.WriteByteToReg(c.Addr, REG_MODE, c.D)
}

func (c *PCAL9535A) GetPin(pin uint) bool {
	return (((c.D >> pin) & 1) == 0)
}
