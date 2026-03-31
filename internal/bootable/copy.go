package bootable

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ilkin0/yaz/internal/progress"
)

const tempISOPath = "/tmp/yaz-iso"

// copyISOContents mounts FAT32 partition and ISO image,
// copies files from ISO to the FAT32 partition (not raw write),
// and verifies UEFI boot loader exists.
func copyISOContents(imagePath, partitionPath string, onProgress progress.Func) error {
	os.MkdirAll(tempISOPath, 0o755)

	deviceMountPath, err := mountPartition(partitionPath)
	if err != nil {
		return fmt.Errorf("cannot mount partition: %s, %w", partitionPath, err)
	}
	defer unmountPartition(partitionPath)

	if err := mountISO(imagePath); err != nil {
		return fmt.Errorf("cannot mount ISO image: %s, %w", imagePath, err)
	}
	defer unmountISO()

	// Calculate total size for progress tracking
	var totalBytes uint64
	filepath.Walk(tempISOPath, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalBytes += uint64(info.Size())
		}
		return nil
	})

	onProgress(progress.Update{LogMessage: "Copying files to USB..."})
	var copiedBytes uint64
	if err := copyDir(tempISOPath, deviceMountPath, func(n uint64) {
		copiedBytes += n
		onProgress(progress.Update{
			Phase:        progress.PhaseWriting,
			BytesWritten: copiedBytes,
			TotalBytes:   totalBytes,
		})
	}); err != nil {
		return fmt.Errorf("copying files: %w", err)
	}

	onProgress(progress.Update{LogMessage: "Verifying UEFI boot loader..."})
	uefiLoader := filepath.Join(deviceMountPath, "EFI", "BOOT", "BOOTX64.EFI")
	if _, err := os.Stat(uefiLoader); os.IsNotExist(err) {
		return fmt.Errorf("UEFI fallback boot loader not found at %s", uefiLoader)
	}

	return nil
}

func copyFile(src, dst string, onBytes func(uint64)) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error open src %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error open dst %w", err)
	}
	defer out.Close()

	buf := make([]byte, 256*1024)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			onBytes(uint64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return out.Sync()
}

func copyDir(src, dst string, onBytes func(uint64)) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, onBytes)
	})
}
