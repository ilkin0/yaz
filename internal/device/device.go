package device

import "fmt"

type (
	DiskType int
	// ByteSize uint64
)

const (
	File DiskType = iota
	Partition
	Disk
)

const LINUX_DVC_PATH = "/sys/block/"

type Block struct {
	Name         string
	DevNode      string
	BlockSize    uint64
	BlockType    DiskType
	Model        string
	PhysicalSize uint64
	Removeable   bool
}

func (b Block) FilterValue() string {
	return b.Name
}

func (b Block) Title() string {
	return b.DevNode
}

func (b Block) Description() string {
	return fmt.Sprintf("%s - %.1f GB", b.Model, float64(b.BlockSize)/(1024*1024*1024))
}
