package writer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ilkin0/yaz/internal/bootable"
	"github.com/ilkin0/yaz/internal/config"
)

const (
	writeBufferSize = 4 * 1024 * 1024
)

func Flash(device, image string, opts config.Options, onProgress ProgressFunc) error {
	f, err := os.Open(image)
	if err != nil {
		return errors.New("cannot open file: " + image)
	}
	defer f.Close()

	iMeta, err := bootable.GetImageMetadata(f)
	if err != nil {
		return fmt.Errorf("reading image metadata: %w", err)
	}
	onProgress(Progress{LogMessage: fmt.Sprintf("ISO is hybrid: %t", iMeta.IsHybrid())})

	if !iMeta.IsHybrid() {
		onProgress(Progress{LogMessage: fmt.Sprintf("Partitioning device: [%s] label: [%s]", device, iMeta.Label)})
		return bootable.MakeUEFIBootDevice(device, iMeta.Label, image, func(msg string, bytesCopied, totalBytes uint64) {
			p := Progress{LogMessage: msg}
			if totalBytes > 0 {
				p.Phase = PhaseWriting
				p.BytesWritten = bytesCopied
				p.TotalBytes = totalBytes
			}
			onProgress(p)
		})
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("cannot seek image file: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return errors.New("cannot open file info for " + image)
	}
	fsize := fi.Size()

	// Unmount device and its partitions before writing
	if onProgress != nil {
		onProgress(Progress{LogMessage: "Unmounting device..."})
	}
	unmountDevice(device)

	if onProgress != nil {
		onProgress(Progress{LogMessage: "Opening device for writing..."})
	}
	dFlags := os.O_WRONLY
	if opts.SyncMode {
		dFlags |= os.O_SYNC
	}
	d, err := os.OpenFile(device, dFlags, 0)
	if err != nil {
		return fmt.Errorf("cannot open device %s: %w", device, err)
	}

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
			// TODO macOS 'rdisk' writes is faster than 'disk' writes
			_, writeErr := d.Write(buff[:n])
			if writeErr != nil {
				d.Close()
				return fmt.Errorf("error writing to device: %w", writeErr)
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
			d.Close()
			return err
		}
	}

	// flush all cached writes to the device
	if onProgress != nil {
		onProgress(Progress{LogMessage: "Syncing device..."})
	}
	if err := syncAndClose(d, device); err != nil {
		return err
	}

	// re-read the partition table
	rereadPartition(device)

	if opts.VerifyWrite {
		if onProgress != nil {
			onProgress(Progress{LogMessage: "Starting verification..."})
		}
		verified, err := Verify(device, image, onProgress)
		if err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		if !verified {
			return errors.New("file integrity failed")
		}
		if onProgress != nil {
			onProgress(Progress{LogMessage: "Verification passed"})
		}
	}
	return nil
}
