package bootable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ilkin0/yaz/internal/progress"
)

const (
	fat32MaxFileSize = 4 * 1024 * 1024 * 1024
	wimSplitMaxMB    = "3800"
)

func MakeUEFIBootDevice(deviceNode, label, imagePath string, onProgress progress.Func) error {
	onProgress(progress.Update{LogMessage: "Analyzing ISO image..."})

	if err := mountISO(imagePath); err != nil {
		return fmt.Errorf("cannot mount ISO image: %s, %w", imagePath, err)
	}
	defer unmountISO()

	// Detect oversized WIM files that need splitting
	var oversizedWIMs []string
	filepath.Walk(tempISOPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Size() >= fat32MaxFileSize {
			if strings.HasSuffix(strings.ToLower(info.Name()), ".wim") {
				oversizedWIMs = append(oversizedWIMs, path)
			}
		}
		return nil
	})

	if len(oversizedWIMs) > 0 {
		if _, err := exec.LookPath("wimlib-imagex"); err != nil {
			return fmt.Errorf("ISO contains files >4GB that must be split for FAT32.\n" +
				"Install wimlib:\n" +
				"  Debian/Ubuntu: sudo apt install wimtools\n" +
				"  Fedora:        sudo dnf install wimlib-utils\n" +
				"  macOS:         brew install wimlib")
		}
	}

	onProgress(progress.Update{LogMessage: "Creating FAT32 partition..."})
	p, err := createSinglePartition(deviceNode, label)
	if err != nil {
		return err
	}
	onProgress(progress.Update{LogMessage: fmt.Sprintf("Partition created: %s", p)})

	if err := copyISOContents(p, oversizedWIMs, onProgress); err != nil {
		return fmt.Errorf("error during ISO copying: %w", err)
	}

	onProgress(progress.Update{LogMessage: "UEFI boot device ready"})
	return nil
}
