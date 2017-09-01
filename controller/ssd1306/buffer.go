package ssd1306

import "errors"

// Buffer abstracts "drawing" into an 8-row page space for Display calls into an SSD1306.
// A suitably-sized instance is created using NewBuffer on an SSD1306 instance.
type Buffer interface {
	On(x, y int) error
	Off(x, y int) error
	Set(x, y int, on bool) error
	FillRect(x, y int, w, h int) error
	ClearRect(x, y int, w, h int) error
	Cells() []byte
}

// bufferHoriz is a Buffer that operates in memory mode of SSD1306_MEMORYMODE_HORIZ
type bufferHoriz struct {
	cells []byte
	width uint
	pages uint
}

func newBuffer(width, pages uint, memoryMode int) Buffer {
	switch memoryMode {
	case SSD1306_MEMORYMODE_HORIZ:
		return newBufferHoriz(width, pages)
	}

	return nil
}

func newBufferHoriz(width, pages uint) *bufferHoriz {
	return &bufferHoriz{
		width: width,
		pages: pages,
		cells: make([]byte, width*pages),
	}
}

func (b *bufferHoriz) Cells() []byte {
	return b.cells
}

func (b *bufferHoriz) On(x, y int) error {
	return b.Set(x, y, true)
}

func (b *bufferHoriz) Off(x, y int) error {
	return b.Set(x, y, false)
}

func (b *bufferHoriz) Set(x, y int, on bool) error {
	if uint(x) > b.width {
		return errors.New("x cannot be greater than buffer width")
	}
	page := uint(y) >> 3
	if page > b.pages {
		return errors.New("y cannot be greater than buffer height")
	}

	index := uint(page*b.width) + uint(x)
	cell := b.cells[index]

	bit := byte(1) << (uint(y) & 0x7)
	if on {
		cell |= bit
	} else {
		cell &^= bit
	}

	b.cells[index] = cell

	return nil
}

func (b *bufferHoriz) FillRect(x, y int, w, h int) error {
	for xi := 0; xi < w; xi++ {
		for yi := 0; yi < h; yi++ {
			if err := b.Set(x+xi, y+yi, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *bufferHoriz) ClearRect(x, y int, w, h int) error {
	for xi := 0; xi < w; xi++ {
		for yi := 0; yi < h; yi++ {
			if err := b.Set(x+xi, y+yi, false); err != nil {
				return err
			}
		}
	}

	return nil
}
