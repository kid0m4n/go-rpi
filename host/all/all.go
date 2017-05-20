// Package all conviniently loads all the inbuilt/supported host drivers.
package all

import (
	_ "github.com/cfreeman/embd/host/bbb"
	_ "github.com/cfreeman/embd/host/edison"
	_ "github.com/cfreeman/embd/host/rpi"
)
