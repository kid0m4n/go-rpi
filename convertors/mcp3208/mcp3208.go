// Package mcp3208 allows interfacing with the mcp3208 8-channel, 12-bit ADC through SPI protocol.
package mcp3208

import (
	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

// MCP3208 represents a mcp3208 8bit DAC.
type MCP3208 struct {
	Mode byte
	Bus  embd.SPIBus
}

const (
	// SingleMode represents the single-ended mode for the mcp3208.
	SingleMode = 1

	// DifferenceMode represents the diffenrential mode for the mcp3208.
	DifferenceMode = 0
)

// New creates a representation of the mcp3208 convertor
func New(mode byte, bus embd.SPIBus) *MCP3208 {
	return &MCP3208{mode, bus}
}

const (
	startBit = 1
)

// AnalogValueAt returns the analog value at the given channel of the convertor.
func (m *MCP3208) AnalogValueAt(chanNum int) (int, error) {
	var data [3]uint8
	data[0] = (uint8(startBit) << 2) + (uint8(m.Mode) << 1) + (uint8(chanNum) >> 2)
	data[1] = uint8(chanNum) << 6
	data[2] = 0

	glog.V(2).Infof("mcp3208: sendingdata buffer %v", data)
	if err := m.Bus.TransferAndReceiveData(data[:]); err != nil {
		return 0, err
	}

	return int(uint16(data[1] & 0x0f) << 8 | uint16(data[2])), nil
}
