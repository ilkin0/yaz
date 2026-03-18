package app

import (
	"github.com/ilkin0/yaz/internal/device"
)

func Run() {
	_, err := device.EnumurateDevice()
	if err != nil {
		panic(err)
	}
}
