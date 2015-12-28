// Package all conviniently loads all the inbuilt/supported host drivers.
package all

import (
	_ "github.com/wujiang/embd/host/bbb"
	_ "github.com/wujiang/embd/host/rpi"
)
