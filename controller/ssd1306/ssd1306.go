/*
Package ssd1306 allows controlling an SSD1306 OLED controller.

This currently supports only write-only operations and a SPI connection to the controller.

Resources

This library is based on these prior implementations:
	https://github.com/adafruit/Adafruit_Python_SSD1306/blob/master/Adafruit_SSD1306/SSD1306.py
	https://github.com/kakaryan/i2cssd1306/blob/master/ssd1306.go

Datasheet
	https://cdn-shop.adafruit.com/datasheets/SSD1306.pdf
*/
package ssd1306

import (
	"github.com/golang/glog"
	"github.com/kidoman/embd"
	"time"
)

const (
	SSD1306_I2C_ADDRESS         = 0x3C
	SSD1306_SETCONTRAST         = 0x81
	SSD1306_DISPLAYALLON_RESUME = 0xA4
	SSD1306_DISPLAYALLON        = 0xA5
	SSD1306_NORMALDISPLAY       = 0xA6
	SSD1306_INVERTDISPLAY       = 0xA7
	SSD1306_DISPLAYOFF          = 0xAE
	SSD1306_DISPLAYON           = 0xAF
	SSD1306_SETDISPLAYOFFSET    = 0xD3
	SSD1306_SETCOMPINS          = 0xDA
	SSD1306_SETVCOMDETECT       = 0xDB
	SSD1306_SETDISPLAYCLOCKDIV  = 0xD5
	SSD1306_SETPRECHARGE        = 0xD9
	SSD1306_SETMULTIPLEX        = 0xA8
	SSD1306_SETLOWCOLUMN        = 0x00
	SSD1306_SETHIGHCOLUMN       = 0x10
	SSD1306_SETSTARTLINE        = 0x40
	SSD1306_MEMORYMODE          = 0x20
	SSD1306_MEMORYMODE_HORIZ    = 0x00
	SSD1306_COLUMNADDR          = 0x21
	SSD1306_PAGEADDR            = 0x22
	SSD1306_COMSCANINC          = 0xC0
	SSD1306_COMSCANDEC          = 0xC8
	SSD1306_SEGREMAP            = 0xA0
	SSD1306_CHARGEPUMP          = 0x8D
	SSD1306_EXTERNALVCC         = 0x1
	SSD1306_SWITCHCAPVCC        = 0x2

	SSD1306_ACTIVATE_SCROLL                      = 0x2F
	SSD1306_DEACTIVATE_SCROLL                    = 0x2E
	SSD1306_SET_VERTICAL_SCROLL_AREA             = 0xA3
	SSD1306_RIGHT_HORIZONTAL_SCROLL              = 0x26
	SSD1306_LEFT_HORIZONTAL_SCROLL               = 0x27
	SSD1306_VERTICAL_AND_RIGHT_HORIZONTAL_SCROLL = 0x29
	SSD1306_VERTICAL_AND_LEFT_HORIZONTAL_SCROLL  = 0x2A
)

const (
	memoryMode = SSD1306_MEMORYMODE_HORIZ
)

// SSD1306 represents an instance of an SSD1306 OLED controller.
type SSD1306 struct {
	spiBus   embd.SPIBus
	dcPin    embd.DigitalPin
	resetPin embd.DigitalPin
	vccState byte
	width    uint
	height   uint
	pages    uint
}

// NewSPI creates a new SSD1306 controller connected via the given SPIBus.
// The GPIO digital output pins that are connected to DC and Rst must also be provided.
// Finally the width x height of the OLED must be given where the width is usually 128 and height is either 32 or 64.
func NewSPI(spiBus embd.SPIBus, dcPin, resetPin embd.DigitalPin, width, height uint) (*SSD1306, error) {
	controller := &SSD1306{
		spiBus:   spiBus,
		dcPin:    dcPin,
		resetPin: resetPin,
		vccState: SSD1306_SWITCHCAPVCC,
		width:    width,
		height:   height,
		pages:    height / 8,
	}

	err := controller.reset()
	if err != nil {
		glog.Errorf("ssd1306: failed to reset: %s", err)
		return nil, err
	}
	err = controller.init()
	if err != nil {
		glog.Errorf("ssd1306: failed to init: %s", err)
		return nil, err
	}

	return controller, nil
}

func (c *SSD1306) reset() error {

	if err := c.resetPin.Write(embd.High); err != nil {
		return err
	}
	time.Sleep(1 * time.Millisecond)

	if err := c.resetPin.Write(embd.Low); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	if err := c.resetPin.Write(embd.High); err != nil {
		return err
	}

	return nil
}

func (c *SSD1306) init() error {

	if err := c.command(SSD1306_DISPLAYOFF); err != nil {
		return err
	}
	if err := c.command(SSD1306_SETDISPLAYCLOCKDIV, 0x80); err != nil {
		return err
	}
	if err := c.command(SSD1306_SETMULTIPLEX, 0x3F); err != nil {
		return err
	}
	if err := c.command(SSD1306_SETDISPLAYOFFSET, 0x0); err != nil {
		return err
	}
	if err := c.command(SSD1306_SETSTARTLINE | 0x0); err != nil {
		return err
	}
	if c.vccState == SSD1306_EXTERNALVCC {
		if err := c.command(SSD1306_CHARGEPUMP, 0x10); err != nil {
			return err
		}
	} else {
		if err := c.command(SSD1306_CHARGEPUMP, 0x14); err != nil {
			return err
		}
	}

	if err := c.command(SSD1306_MEMORYMODE, memoryMode); err != nil {
		return err
	}
	if err := c.command(SSD1306_SEGREMAP | 0x1); err != nil {
		return err
	}
	if err := c.command(SSD1306_COMSCANDEC); err != nil {
		return err
	}
	if err := c.command(SSD1306_SETCOMPINS, 0x12); err != nil {
		return err
	}
	if c.vccState == SSD1306_EXTERNALVCC {
		if err := c.command(SSD1306_SETCONTRAST, 0x9F); err != nil {
			return err
		}
	} else {
		if err := c.command(SSD1306_SETCONTRAST, 0xCF); err != nil {
			return err
		}
	}

	if c.vccState == SSD1306_EXTERNALVCC {
		if err := c.command(SSD1306_SETPRECHARGE, 0x22); err != nil {
			return err
		}

	} else {
		if err := c.command(SSD1306_SETPRECHARGE, 0xF1); err != nil {
			return err
		}
	}

	if err := c.command(SSD1306_SETVCOMDETECT, 0x40); err != nil {
		return err
	}
	if err := c.command(SSD1306_DISPLAYALLON_RESUME); err != nil {
		return err
	}
	if err := c.command(SSD1306_NORMALDISPLAY); err != nil {
		return err
	}

	if err := c.command(SSD1306_DISPLAYON); err != nil {
		return err
	}

	return nil
}

func (c *SSD1306) command(cmd ...byte) error {
	c.dcPin.Write(embd.Low)
	_, err := c.spiBus.Write(cmd)

	return err
}

func (c *SSD1306) data(d ...byte) error {
	c.dcPin.Write(embd.High)
	_, err := c.spiBus.Write(d)

	return err
}

// Display sends the given buffer to the controller to "rendered"
func (c *SSD1306) Display(buf Buffer) error {
	if err := c.command(SSD1306_COLUMNADDR, 0, byte(c.width-1)); err != nil {
		return err
	}
	if err := c.command(SSD1306_PAGEADDR, 0, byte(c.pages-1)); err != nil {
		return err
	}
	return c.data(buf.Cells()...)
}

// Close turns the display off
func (c *SSD1306) Close() error {
	if err := c.command(SSD1306_DISPLAYOFF); err != nil {
		return err
	}
	return nil
}

// NewBuffer creates a buffer that is suitably configured to be used in Display calls.
func (c *SSD1306) NewBuffer() Buffer {
	return newBuffer(c.width, c.pages, memoryMode)
}
