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

// +build ignore

package main

import (
	"fmt"

	"github.com/cfreeman/embd"
	_ "github.com/cfreeman/embd/host/all"
)

func main() {
	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()

	spiBus := embd.NewSPIBus(embd.SPIMode0, 0, 1000000, 8, 0)
	defer spiBus.Close()

	dataBuf := [3]uint8{1, 2, 3}

	if err := spiBus.TransferAndReceiveData(dataBuf[:]); err != nil {
		panic(err)
	}

	fmt.Println("received data is:", dataBuf)

	dataReceived, err := spiBus.ReceiveData(3)
	if err != nil {
		panic(err)
	}

	fmt.Println("received data is:", dataReceived)

	dataByte := byte(1)
	receivedByte, err := spiBus.TransferAndReceiveByte(dataByte)
	if err != nil {
		panic(err)
	}
	fmt.Println("received byte is:", receivedByte)

	receivedByte, err = spiBus.ReceiveByte()
	if err != nil {
		panic(err)
	}
	fmt.Println("received byte is:", receivedByte)
}
