package writer

type Progress struct {
	BytesWritten uint64
	TotalBytes   uint64
	Speed        float64
}

type ProgressFunc func(Progress)
