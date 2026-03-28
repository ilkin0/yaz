package writer

import (
	"os"
	"os/exec"
)

func unmountDevice(device string) {
	exec.Command("diskutil", "unmountDisk", "force", device).Run()
}

func syncAndClose(d *os.File, device string) error {
	// macOS doesn't support fsync on raw block devices;
	// close the fd first, then unmountDisk flushes the buffers
	d.Close()
	exec.Command("diskutil", "unmountDisk", device).Run()
	return nil
}

func rereadPartition(device string) {
	exec.Command("diskutil", "mountDisk", device).Run()
}
