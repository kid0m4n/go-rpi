// Package mcp3008 allows interfacing with the mcp3008 8-channel, 10-bit ADC through SPI protocol.
package mcp3008

import (
	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

// MCP3008 represents a mcp3008 8bit DAC.
type MCP3008 struct {
	Mode, Bits byte
	Bus        embd.SPIBus
}

//How many bits does the
const (
	Bits10 = iota //10 bit MCP300** family
	Bits12        //12 bit MCP320** family
)

const (
	// SingleMode represents the single-ended mode for the mcp3008.
	SingleMode = 1

	// DifferenceMode represents the diffenrential mode for the mcp3008.
	DifferenceMode = 0
)

// New creates a representation of the mcp3008 convertor
func New(mode, bits byte, bus embd.SPIBus) *MCP3008 {
	return &MCP3008{Mode: mode, Bus: bus}
}

const (
	startBit = 1
)

// AnalogValueAt returns the analog value at the given channel of the convertor.
func (m *MCP3008) AnalogValueAt(chanNum int) (int, error) {
	var data [3]uint8
	switch m.Bits {
	case Bits10:
		data[0] = startBit
		data[1] = uint8(m.Mode)<<7 | uint8(chanNum)<<4
		data[2] = 0
	case Bits12:
		data[0] = (uint8(startBit) << 2) + (uint8(m.Mode) << 1) + (uint8(chanNum) >> 2)
		data[1] = uint8(chanNum) << 6
		data[2] = 0
	}

	glog.V(2).Infof("mcp3008: sendingdata buffer %v", data)
	if err := m.Bus.TransferAndReceiveData(data[:]); err != nil {
		return 0, err
	}
	switch m.Bits {
	case Bits10:
		return int(uint16(data[1]&0x03)<<8 | uint16(data[2])), nil
	case Bits12:
		return int(uint16(data[1]&0x0f)<<8 | uint16(data[2])), nil
	default:
		panic("mcp3008: unknown number of bits")
	}

}
