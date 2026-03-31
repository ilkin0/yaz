package bootable

import (
	"fmt"

	"github.com/ilkin0/yaz/internal/progress"
)

func MakeUEFIBootDevice(deviceNode, label, imagePath string, onProgress progress.Func) error {
	onProgress(progress.Update{LogMessage: "Creating GPT partition table..."})
	p, err := createGptPartition(deviceNode, label)
	if err != nil {
		return err
	}
	onProgress(progress.Update{LogMessage: fmt.Sprintf("Partition created: %s", p)})

	onProgress(progress.Update{LogMessage: "Mounting partition and ISO..."})
	if err := copyISOContents(imagePath, p, onProgress); err != nil {
		return fmt.Errorf("error during ISO copying: %w", err)
	}

	onProgress(progress.Update{LogMessage: "UEFI boot device ready"})
	return nil
}
