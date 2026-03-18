package device

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
	name         string
	devNode      string
	blockSize    uint64
	blockType    DiskType
	model        string
	physicalSize uint64
	removeable   bool
}
