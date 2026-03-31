package bootable

import "fmt"

// ProgressFunc reports copy progress to the TUI.
// If msg is non-empty, it's a log message.
// If totalBytes > 0, it's a byte-level progress update.
type ProgressFunc func(msg string, bytesCopied, totalBytes uint64)

func MakeUEFIBootDevice(deviceNode, label, imagePath string, onProgress ProgressFunc) error {
	onProgress("Creating GPT partition table...", 0, 0)
	p, err := createGptPartition(deviceNode, label)
	if err != nil {
		return err
	}
	onProgress(fmt.Sprintf("Partition created: %s", p), 0, 0)

	onProgress("Mounting partition and ISO...", 0, 0)
	if err := copyISOContents(imagePath, p, onProgress); err != nil {
		return fmt.Errorf("error during ISO copying: %w", err)
	}

	onProgress("UEFI boot device ready", 0, 0)
	return nil
}
