package writer

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ilkin0/yaz/internal/bootable"
	"github.com/ilkin0/yaz/internal/config"
	"github.com/ilkin0/yaz/internal/progress"
)

const (
	writeBufferSize = 4 * 1024 * 1024
)

func Flash(device, image string, opts config.Options, onProgress progress.Func) error {
	f, err := os.Open(image)
	if err != nil {
		return errors.New("cannot open file: " + image)
	}
	defer f.Close()

	iMeta, err := bootable.GetImageMetadata(f)
	if err != nil {
		return fmt.Errorf("reading image metadata: %w", err)
	}
	onProgress(progress.Update{LogMessage: fmt.Sprintf("ISO is hybrid: %t", iMeta.IsHybrid())})

	if !iMeta.IsHybrid() {
		onProgress(progress.Update{LogMessage: fmt.Sprintf("Partitioning device: [%s] label: [%s]", device, iMeta.Label)})
		return bootable.MakeUEFIBootDevice(device, iMeta.Label, image, onProgress)
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
	onProgress(progress.Update{LogMessage: "Unmounting device..."})
	unmountDevice(device)

	onProgress(progress.Update{LogMessage: "Opening device for writing..."})
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

			onProgress(progress.Update{
				Phase:        progress.PhaseWriting,
				BytesWritten: written,
				TotalBytes:   uint64(fsize),
			})
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
	onProgress(progress.Update{LogMessage: "Syncing device..."})
	if err := syncAndClose(d, device); err != nil {
		return err
	}

	// re-read the partition table
	rereadPartition(device)

	if opts.VerifyWrite {
		onProgress(progress.Update{LogMessage: "Starting verification..."})
		verified, err := Verify(device, image, onProgress)
		if err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		if !verified {
			return errors.New("file integrity failed")
		}
		onProgress(progress.Update{LogMessage: "Verification passed"})
	}
	return nil
}
