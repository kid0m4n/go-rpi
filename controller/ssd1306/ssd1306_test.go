package ssd1306

import (
	"github.com/kidoman/embd"
	"testing"
	"time"
)

type mockSpiBus struct {
	chunks [][]byte
}

func (s *mockSpiBus) Write(p []byte) (n int, err error) {
	s.chunks = append(s.chunks, p)
	return 0, nil
}
func (s *mockSpiBus) TransferAndReceiveData(dataBuffer []uint8) error { return nil }
func (s *mockSpiBus) ReceiveData(len int) ([]uint8, error)            { return nil, nil }
func (s *mockSpiBus) TransferAndReceiveByte(data byte) (byte, error)  { return 0, nil }
func (s *mockSpiBus) ReceiveByte() (byte, error)                      { return 0, nil }
func (s *mockSpiBus) Close() error                                    { return nil }

type mockPin struct {
	values []int
}

func (p *mockPin) Watch(edge embd.Edge, handler func(embd.DigitalPin)) error { return nil }
func (p *mockPin) StopWatching() error                                       { return nil }
func (p *mockPin) N() int                                                    { return 0 }
func (p *mockPin) Write(val int) error {
	p.values = append(p.values, val)
	return nil
}
func (p *mockPin) Read() (int, error)                         { return 0, nil }
func (p *mockPin) TimePulse(state int) (time.Duration, error) { return 0, nil }
func (p *mockPin) SetDirection(dir embd.Direction) error      { return nil }
func (p *mockPin) ActiveLow(b bool) error                     { return nil }
func (p *mockPin) PullUp() error                              { return p.Write(1) }
func (p *mockPin) PullDown() error                            { return p.Write(0) }
func (p *mockPin) Close() error                               { return nil }

func TestInit(t *testing.T) {

	spiBus := &mockSpiBus{}
	dcPin := &mockPin{}
	resetPin := &mockPin{}

	controller, err := NewSPI(spiBus, dcPin, resetPin, 128, 64)
	if err != nil {
		t.Fatalf("Shouldn't be an error: %s", err)
	}
	if controller == nil {
		t.Fatal("controller shouldn't be nil")
	}

	if len(resetPin.values) != 3 {
		t.Error("expected 3 touches to the reset pin")
	}

	if len(dcPin.values) != 16 {
		t.Error("expected 16 touches to the dc pin")
	}

	if len(spiBus.chunks) != 16 {
		t.Error("expected 16 commands during init")
	}
	i := 0

	if spiBus.chunks[i][0] != SSD1306_DISPLAYOFF {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETDISPLAYCLOCKDIV {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETMULTIPLEX {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETDISPLAYOFFSET {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETSTARTLINE|0x0 {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_CHARGEPUMP {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_MEMORYMODE {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SEGREMAP|0x1 {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_COMSCANDEC {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETCOMPINS {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETCONTRAST {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETPRECHARGE {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_SETVCOMDETECT {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_DISPLAYALLON_RESUME {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_NORMALDISPLAY {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

	if spiBus.chunks[i][0] != SSD1306_DISPLAYON {
		t.Errorf("Wrong command in chunk %d", i)
	}
	i++

}
