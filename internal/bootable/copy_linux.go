package bootable

import (
	"os"
	"os/exec"
)

const tempDevicePath = "/tmp/yaz-usb"

func mountPartition(partition string) (string, error) {
	os.MkdirAll(tempDevicePath, 0o755)
	if err := exec.Command("mount", partition, tempDevicePath).Run(); err != nil {
		return "", err
	}
	return tempDevicePath, nil
}

func mountISO(imagePath string) error {
	return exec.Command("mount", "-o", "loop,ro", imagePath, tempISOPath).Run()
}

func unmountPartition(_ string) {
	exec.Command("umount", "-l", tempDevicePath).Run()
}

func unmountISO() {
	exec.Command("umount", "-l", tempISOPath).Run()
}
