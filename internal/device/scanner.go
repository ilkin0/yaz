package device

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func EnumurateDevice() ([]Block, error) {
	switch runtime.GOOS {
	case "linux":
		return enumrateLinux()
	}

	return nil, errors.New("error for device enumirations")
}

func enumrateLinux() ([]Block, error) {
	c, err := os.ReadDir(linuxSysBlockPath)
	if err != nil {
		return nil, errors.New("error happened during DIR listing")
	}

	var vblocks []Block
	for _, entry := range c {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// skip loop/ram partitions
		if strings.HasPrefix(info.Name(), "loop") || strings.HasPrefix(info.Name(), "ram") {
			continue
		}

		blockPath := linuxSysBlockPath + info.Name()

		data, err := os.ReadFile(blockPath + "/removable")
		if err != nil {
			continue
		}
		isRemovable := strings.TrimSpace(string(data)) == "1"
		if !isRemovable {
			continue
		}

		data, err = os.ReadFile(blockPath + "/size")
		if err != nil {
			continue
		}
		sizeVal, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		if err != nil {
			continue
		}

		data, err = os.ReadFile(blockPath + "/device/model")
		if err != nil {
			continue
		}
		bModel := strings.TrimSpace(string(data))

		vblocks = append(vblocks, Block{
			Name:       info.Name(),
			Removeable: isRemovable,
			BlockSize:  (sizeVal * 512), // converting to bytes (sizeVal * 512), sysfs is in 512-byte sectors
			DevNode:    "/dev/" + info.Name(),
			Model:      bModel,
		})
	}

	return vblocks, nil
}
