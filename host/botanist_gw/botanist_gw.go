/*
        Package botaist_gw provides Botanist GW support.
        The following features are supported on Linux kernel 3.8+

        GPIO (digital (rw))
        I<C2><B2>C
        LED
*/
package botanist_gw

import (
        "github.com/kidoman/embd"
        "github.com/kidoman/embd/host/generic"
)

const BotanistGwHost = "botanist_gw"

var spiDeviceMinor = int(32766)

var pins = embd.PinMap{
	&embd.PinDesc{ID: "PA27", Aliases: []string{"27", "PORT_ENABLE", }, Caps: embd.CapDigital, DigitalLogical: 27},
	&embd.PinDesc{ID: "PA28", Aliases: []string{"28", "PORT_FAULT", }, Caps: embd.CapDigital, DigitalLogical: 28},
	&embd.PinDesc{ID: "PA29", Aliases: []string{"29", "PORT_DETECT", }, Caps: embd.CapDigital, DigitalLogical: 29},
}

var ledMap = embd.LEDMap{
}

func init() {
        embd.Register(BotanistGwHost, func(rev int) *embd.Descriptor {
                return &embd.Descriptor{
                        GPIODriver: func() embd.GPIODriver {
                                return embd.NewGPIODriver(pins, NewDigitalPin, nil, nil)
                        },
                        I2CDriver: func() embd.I2CDriver {
                                return embd.NewI2CDriver(generic.NewI2CBus)
                        },
                        SPIDriver: func() embd.SPIDriver {
                                return embd.NewSPIDriver(spiDeviceMinor, generic.NewSPIBus, nil)
                        },
                }
        })
}
