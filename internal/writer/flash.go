package writer

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ilkin0/yaz/internal/config"
)

const (
	DD_BUFFER_SIZE int = 4
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

	buff := make([]byte, DD_BUFFER_SIZE*1024*1024)
	var written uint64
	lastTime := time.Now()
	var lastWritten uint64
	var smoothSpeed float64
	const smoothing = 0.3 // weight for new speed sample (lower = smoother)

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
			if elapsed > 0.5 { // update speed every 500ms minimum
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

	return nil
}
