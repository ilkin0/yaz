package bootable

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ilkin0/yaz/internal/progress"
)

const tempISOPath = "/tmp/yaz-iso"

// copyISOContents mounts a FAT32 partition, copies files from the mounted ISO,
// splits any oversized WIM files into <4GB chunks, and verifies the UEFI boot loader.
func copyISOContents(partitionPath string, oversizedWIMs []string, onProgress progress.Func) error {
	deviceMountPath, err := mountPartition(partitionPath)
	if err != nil {
		return fmt.Errorf("cannot mount partition: %s, %w", partitionPath, err)
	}
	defer unmountPartition(partitionPath)

	// Build a set of oversized WIM paths for quick lookup
	skipFiles := make(map[string]bool, len(oversizedWIMs))
	for _, w := range oversizedWIMs {
		skipFiles[w] = true
	}

	var totalBytes uint64
	filepath.Walk(tempISOPath, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalBytes += uint64(info.Size())
		}
		return nil
	})

	onProgress(progress.Update{LogMessage: fmt.Sprintf("Total: %d files, %s", fileCount(tempISOPath), progress.HumanBytes(totalBytes))})
	onProgress(progress.Update{LogMessage: "Copying files to USB..."})

	var copiedBytes uint64
	onBytes := func(n uint64) {
		copiedBytes += n
		onProgress(progress.Update{
			Phase:        progress.PhaseWriting,
			BytesWritten: copiedBytes,
			TotalBytes:   totalBytes,
		})
	}

	// Copy all files, skipping oversized WIMs (they'll be split separately)
	if err := filepath.Walk(tempISOPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(tempISOPath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(deviceMountPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		if skipFiles[path] {
			onProgress(progress.Update{LogMessage: fmt.Sprintf("Skipping %s (%s) — will split", relPath, progress.HumanBytes(uint64(info.Size())))})
			return nil
		}

		return copyFile(path, targetPath, onBytes)
	}); err != nil {
		return fmt.Errorf("copying files: %w", err)
	}

	// Split oversized WIM files into <4GB .swm chunks directly on the USB
	for _, wimPath := range oversizedWIMs {
		relPath, _ := filepath.Rel(tempISOPath, wimPath)
		wimInfo, _ := os.Stat(wimPath)

		onProgress(progress.Update{LogMessage: fmt.Sprintf("Splitting %s (%s) into <4GB chunks...", relPath, progress.HumanBytes(uint64(wimInfo.Size())))})

		// Change .wim → .swm for the split output
		dstPath := filepath.Join(deviceMountPath, relPath)
		dstPath = strings.TrimSuffix(dstPath, filepath.Ext(dstPath)) + ".swm"

		cmd := exec.Command("wimlib-imagex", "split", wimPath, dstPath, wimSplitMaxMB)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("splitting %s: %s, %w", relPath, string(out), err)
		}

		onProgress(progress.Update{LogMessage: fmt.Sprintf("Split %s complete", relPath)})
	}

	onProgress(progress.Update{LogMessage: "Verifying UEFI boot loader..."})
	uefiLoader := filepath.Join(deviceMountPath, "EFI", "BOOT", "BOOTX64.EFI")
	if _, err := os.Stat(uefiLoader); os.IsNotExist(err) {
		return fmt.Errorf("UEFI fallback boot loader not found at %s", uefiLoader)
	}

	return nil
}

func fileCount(dir string) int {
	count := 0
	filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
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

	buf := make([]byte, 1024*1024)
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
