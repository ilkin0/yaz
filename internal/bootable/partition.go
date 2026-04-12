package bootable

import (
	"fmt"
	"os/exec"
	"runtime"
)

// createSinglePartition creates a single FAT32 partition spanning the entire device.
func createSinglePartition(device, label string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("diskutil",
			"partitionDisk",
			device,
			"GPT",
			"FAT32",
			label,
			"100%")

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("partitioning %s: %w", device, err)
		}

		cmd = exec.Command("diskutil", "mountDisk", device)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("mounting %s: %w", device, err)
		}
		return "", nil
	case "linux":
		if err := exec.Command("parted", device, "--script", "mklabel", "gpt").Run(); err != nil {
			return "", fmt.Errorf("creating GPT label on %s: %w", device, err)
		}

		if err := exec.Command("parted", device, "--script", "mkpart", "primary", "fat32", "1MiB", "100%").Run(); err != nil {
			return "", fmt.Errorf("creating partition on %s: %w", device, err)
		}

		partition := device + "1"
		if err := exec.Command("mkfs.vfat", "-F", "32", "-n", label, partition).Run(); err != nil {
			return "", fmt.Errorf("formatting %s as FAT32: %w", partition, err)
		}
		return partition, nil
	}

	return "", nil
}
