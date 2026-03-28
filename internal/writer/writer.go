package writer

type Phase int

const (
	PhaseWriting Phase = iota
	PhaseVerifying
)

type Progress struct {
	Phase        Phase
	BytesWritten uint64
	TotalBytes   uint64
	Speed        float64
	LogMessage   string
}

type ProgressFunc func(Progress)
