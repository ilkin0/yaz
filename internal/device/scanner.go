package device

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func EnumurateDevice() ([]Block, error) {
	switch runtime.GOOS {
	case "linux":
		return enumrateLinux()
	case "darwin":
		return enumrateDarwin()
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

func enumrateDarwin() ([]Block, error) {
	out, err := exec.Command("diskutil", "list").Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil list: %w", err)
	}

	// TODO bufio scanner instead for better performance?
	lines := strings.Split(string(out), "\n")
	var devices []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/dev/disk") && strings.Contains(line, "external") && strings.Contains(line, "physical") {
			parts := strings.SplitN(line, " ", 2)
			devices = append(devices, parts[0])
		}
	}

	var blocks []Block
	for _, d := range devices {
		out, err = exec.Command("diskutil", "info", d).Output()
		if err != nil {
			return nil, fmt.Errorf("diskutil info %s: %w", d, err)
		}
		b := Block{
			DevNode: d,
			Name:    strings.TrimPrefix(d, "/dev/"),
		}

		lines := strings.SplitSeq(string(out), "\n")
		for line := range lines {
			parts := strings.Split(line, ":")
			key := parts[0]
			if len(parts) > 1 {
				value := strings.TrimSpace(parts[1])
				if strings.Contains(key, "Media Name") {
					b.Model = value
				} else if strings.Contains(key, "Removable") {
					b.Removeable = value == "Removable"
				} else if strings.Contains(key, "Disk Size") {
					// format: "Disk Size:  15.7 GB (15664676864 Bytes)..."
					if idx := strings.Index(line, "("); idx != -1 {
						after := line[idx+1:]
						bytesStr, _, _ := strings.Cut(after, " ")
						if n, parseErr := strconv.ParseUint(bytesStr, 10, 64); parseErr == nil {
							b.BlockSize = n
						}
					}
				}
			}
		}
		blocks = append(blocks, b)
	}

	return blocks, nil
}
