package writer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/ilkin0/yaz/internal/config"
)

const (
	writeBufferSize int = 4 * 1024 * 1024
)

func Flash(device, image string, opts config.Options, onProgress ProgressFunc) error {
	f, err := os.Open(image)
	if err != nil {
		return errors.New("cannot open file: " + image)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return errors.New("cannot open file info for " + image)
	}
	fsize := fi.Size()

	// Unmount device and its partitions before open/writing
	devDir := "/dev"
	devicePrefix, _ := strings.CutPrefix(device, devDir+"/")
	var partitions []string
	entries, _ := os.ReadDir(devDir)

	regex := regexp.MustCompile(`^` + regexp.QuoteMeta(devicePrefix) + `\d+$`)
	for _, entry := range entries {
		name := entry.Name()
		if regex.MatchString(name) {
			partitions = append(partitions, devDir+"/"+name)
		}
	}

	for _, p := range partitions {
		exec.Command("umount", "-l", p).Run()
	}

	dFlags := os.O_WRONLY
	if opts.SyncMode {
		dFlags |= os.O_SYNC
	}
	d, err := os.OpenFile(device, dFlags, 0)
	if err != nil {
		return errors.New("cannot open device: " + device)
	}
	defer d.Close()

	if !opts.QuickFormat {
		// TODO zero-fill device first
	}

	buff := make([]byte, writeBufferSize)
	var written uint64
	lastTime := time.Now()
	var lastWritten uint64
	var smoothSpeed float64
	const smoothing = 0.3

	for {
		n, err := f.Read(buff)
		if n > 0 {
			_, writeErr := d.Write(buff[:n])
			if writeErr != nil {
				return errors.New("error happened during image writing")
			}
			written += uint64(n)

			now := time.Now()
			elapsed := now.Sub(lastTime).Seconds()
			if elapsed > 0.5 {
				chunkSpeed := float64(written-lastWritten) / elapsed
				if smoothSpeed == 0 {
					smoothSpeed = chunkSpeed
				} else {
					smoothSpeed = smoothing*chunkSpeed + (1-smoothing)*smoothSpeed
				}
				lastTime = now
				lastWritten = written
			}

			if onProgress != nil {
				onProgress(Progress{
					Phase:        PhaseWriting,
					BytesWritten: written,
					TotalBytes:   uint64(fsize),
					Speed:        smoothSpeed,
				})
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// flush all cached writes to the device
	if err := d.Sync(); err != nil {
		return errors.New("error syncing device: " + err.Error())
	}

	// close device before re-reading partition table
	d.Close()

	// tell kernel to re-read the partition table
	exec.Command("partprobe", device).Run()

	if opts.VerifyWrite {
		verified, err := Verify(device, image, onProgress)
		if err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		if !verified {
			return errors.New("file integrity failed")
		}
	}
	return nil
}
