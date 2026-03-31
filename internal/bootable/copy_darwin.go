package bootable

import (
	"fmt"
	"os/exec"
	"strings"
)

func mountPartition(partition string) (string, error) {
	// diskutil partitionDisk auto-mounts to /Volumes/<label>
	// Find the mount point from diskutil info
	out, err := exec.Command("diskutil", "info", partition).Output()
	if err != nil {
		return "", fmt.Errorf("diskutil info %s: %w", partition, err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Mount Point") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				mountPoint := strings.TrimSpace(parts[1])
				if mountPoint != "" {
					return mountPoint, nil
				}
			}
		}
	}

	return "", fmt.Errorf("partition %s is not mounted", partition)
}

func mountISO(imagePath string) error {
	return exec.Command("hdiutil", "attach", "-readonly", "-mountpoint", tempISOPath, imagePath).Run()
}

func unmountPartition(partition string) {
	exec.Command("diskutil", "unmount", partition).Run()
}

func unmountISO() {
	exec.Command("hdiutil", "detach", tempISOPath).Run()
}
