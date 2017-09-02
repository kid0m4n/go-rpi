package ssd1306

import (
	"reflect"
	"testing"
)

func TestAllocation(t *testing.T) {
	buffer := newBuffer(16, 2, memoryMode)
	if buffer == nil {
		t.Fatal("buffer shouldn't be nil")
	}

	if buffer.Cells() == nil {
		t.Fatal("cells shouldn't be nil")
	}

	if len(buffer.Cells()) != 32 {
		t.Error("wrong cell count")
	}
}

var onTests = []struct {
	name     string
	x, y     int
	expected []byte
}{
	{name: "top left", x: 0, y: 0, expected: []byte{0x1, 0, 0, 0, 0, 0, 0, 0 /*page*/, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "left top of page 2", x: 0, y: 8, expected: []byte{0, 0, 0, 0, 0, 0, 0, 0 /*page*/, 0x1, 0, 0, 0, 0, 0, 0, 0}},
	{name: "left row 2", x: 0, y: 1, expected: []byte{1 << 1, 0, 0, 0, 0, 0, 0, 0 /*page*/, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "middle ish", x: 4, y: 3, expected: []byte{0, 0, 0, 0, 1 << 3, 0, 0, 0 /*page*/, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "bottom ish", x: 6, y: 14, expected: []byte{0, 0, 0, 0, 0, 0, 0, 0 /*page*/, 0, 0, 0, 0, 0, 0, 1 << (14 - 8), 0}},
}

func TestBufferHoriz_On(t *testing.T) {
	width := uint(8)
	pages := uint(2) // height = 16

	for _, tt := range onTests {
		// Disabling testing Run use since this needs to build in go 1.6 also
		//t.Run(tt.name, func(t *testing.T) {
		buffer := newBufferHoriz(width, pages)

		buffer.On(tt.x, tt.y)

		if !reflect.DeepEqual(buffer.cells, tt.expected) {
			t.Errorf("%s: wrong cell content, saw %v", tt.name, buffer.cells)
		}
		//})
	}
}

func TestBufferHoriz_Set(t *testing.T) {
	width := uint(8)
	pages := uint(2) // height = 16

	for _, tt := range onTests {
		// Disabling testing Run use since this needs to build in go 1.6 also
		//t.Run(tt.name, func(t *testing.T) {
		buffer := newBufferHoriz(width, pages)
		buffer.Set(tt.x, tt.y, true)

		if !reflect.DeepEqual(buffer.cells, tt.expected) {
			t.Errorf("%s: wrong cell content, saw %v", tt.name, buffer.cells)
		}
		//})
	}
}

func TestBufferHoriz_FillRect(t *testing.T) {
	width := uint(8)
	pages := uint(2) // height = 16

	tests := []struct {
		name          string
		x, y          int
		width, height int
		expected      []byte
	}{
		{name: "all", x: 0, y: 0, width: 8, height: 16, expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: "top half", x: 0, y: 0, width: 8, height: 8, expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0}},
		{name: "top left", x: 0, y: 0, width: 4, height: 8, expected: []byte{0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		{name: "top quarter", x: 0, y: 0, width: int(width), height: 4, expected: []byte{0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0, 0, 0, 0, 0, 0, 0, 0}},
	}

	for _, tt := range tests {
		// Disabling testing Run use since this needs to build in go 1.6 also
		//t.Run(tt.name, func(t *testing.T) {
		buffer := newBufferHoriz(width, pages)

		buffer.FillRect(tt.x, tt.y, tt.width, tt.height)

		if !reflect.DeepEqual(buffer.cells, tt.expected) {
			t.Errorf("%s: wrong cell content, saw %v", tt.name, buffer.cells)
		}
		//})
	}
}

func TestBufferHoriz_Off(t *testing.T) {
	width := uint(8)
	pages := uint(2) // height = 16

	tests := []struct {
		name     string
		x, y     int
		expected []byte
	}{
		{name: "top left", x: 0, y: 0, expected: []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: "bottom of upper page", x: 1, y: 7, expected: []byte{0xff, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: "top of second page", x: 2, y: 8, expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	for _, tt := range tests {
		// Disabling testing Run use since this needs to build in go 1.6 also
		//t.Run(tt.name, func(t *testing.T) {
		buffer := newBufferHoriz(width, pages)

		// assumes TestBufferHoriz_FillRect passes
		buffer.FillRect(0, 0, 8, 16)

		buffer.Off(tt.x, tt.y)

		if !reflect.DeepEqual(buffer.cells, tt.expected) {
			t.Errorf("%s: wrong cell content, saw %v", tt.name, buffer.cells)
		}

		//})
	}
}
