package bootable

import (
	"os"
	"regexp"
	"strings"
)

type ImageMetadata struct {
	MBRPartition bool
	GPTPartition bool
	Label        string
}

func (i ImageMetadata) IsHybrid() bool {
	return i.GPTPartition || i.MBRPartition
}

// GetImageMetadata valdiates whether ISO image is hybrid or not,
// it checks whether image has MBR or GPT partitions, if not it means we need to create a partition for the image.
//
// MBR details are in the first 512 bytes (510 and 511 is the MBR signature - 0x55AA)
// GPT details are more compact, signature offest is 8
func GetImageMetadata(image *os.File) (ImageMetadata, error) {
	meta := ImageMetadata{}
	// Check MBR partition
	mbr := make([]byte, 512)
	if _, err := image.ReadAt(mbr, 0); err == nil {
		meta.MBRPartition = mbr[510] == 0x55 && mbr[511] == 0xAA
	}

	// Check GPT partition
	gpt := make([]byte, 8)
	if _, err := image.ReadAt(gpt, 512); err == nil {
		meta.GPTPartition = string(gpt) == "EFI PART"
	}

	buff := make([]byte, 2048)           // one sector = 2048
	_, err := image.ReadAt(buff, 0x8000) // Based ISO9660 standart PVD is at offest 0x8000 or 32768 bytes
	if err != nil {
		return ImageMetadata{}, err
	}
	volID := strings.TrimSpace(string(buff[40:72]))
	meta.Label = cleanLabel(volID)

	return meta, nil
}

var labelCleanRe = regexp.MustCompile(`[^A-Z0-9_]`)

func cleanLabel(label string) string {
	label = strings.ToUpper(label)

	// keep only A-Z, 0-9, _
	label = labelCleanRe.ReplaceAllString(label, "")

	// FAT32 (MS-DOS) volume labels must be <= 11 characters
	if len(label) > 11 {
		label = label[:11]
	}

	if label == "" {
		label = "USB"
	}

	return label
}
