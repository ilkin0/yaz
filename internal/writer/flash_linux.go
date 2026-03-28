package writer

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func unmountDevice(device string) {
	devDir := "/dev"
	devicePrefix, _ := strings.CutPrefix(device, devDir+"/")
	entries, _ := os.ReadDir(devDir)
	regex := regexp.MustCompile(`^` + regexp.QuoteMeta(devicePrefix) + `\d+$`)
	for _, entry := range entries {
		if regex.MatchString(entry.Name()) {
			exec.Command("umount", "-l", devDir+"/"+entry.Name()).Run()
		}
	}
}

func syncAndClose(d *os.File, _ string) error {
	defer d.Close()
	if err := d.Sync(); err != nil {
		return fmt.Errorf("syncing device: %w", err)
	}
	return nil
}

func rereadPartition(device string) {
	exec.Command("partprobe", device).Run()
}
