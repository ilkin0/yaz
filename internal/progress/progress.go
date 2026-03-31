package progress

type Phase int

const (
	PhaseWriting Phase = iota
	PhaseVerifying
)

type Update struct {
	Phase        Phase
	BytesWritten uint64
	TotalBytes   uint64
	LogMessage   string
}

type Func func(Update)
