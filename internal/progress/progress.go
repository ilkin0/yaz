package progress

import "fmt"

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

// HumanBytes formats a byte count into a human-readable string (e.g. "1.5 GB").
func HumanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
