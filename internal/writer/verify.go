package writer

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/ilkin0/yaz/internal/progress"
)

const verifyBufferSize = 4 * 1024 * 1024

func Verify(device, filePath string, onProgress progress.Func) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, errors.New("failed to open file: " + filePath)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return false, errors.New("failed to stat file: " + filePath)
	}
	totalBytes := uint64(fi.Size())

	d, err := os.Open(device)
	if err != nil {
		return false, errors.New("cannot open device: " + device)
	}
	defer d.Close()

	fileBuf := make([]byte, verifyBufferSize)
	deviceBuf := make([]byte, verifyBufferSize)
	var verified uint64

	for {
		fn, ferr := io.ReadFull(file, fileBuf)
		dn, derr := io.ReadFull(d, deviceBuf[:fn])

		if fn > 0 {
			if dn != fn {
				return false, errors.New("device read returned fewer bytes than image")
			}

			if !bytes.Equal(fileBuf[:fn], deviceBuf[:fn]) {
				return false, nil
			}

			verified += uint64(fn)

			onProgress(progress.Update{
				Phase:        progress.PhaseVerifying,
				BytesWritten: verified,
				TotalBytes:   totalBytes,
			})
		}

		if ferr != nil {
			if ferr == io.EOF || ferr == io.ErrUnexpectedEOF {
				break
			}
			return false, errors.New("failed to read file: " + filePath)
		}

		if derr != nil && derr != io.ErrUnexpectedEOF {
			return false, errors.New("failed to read device: " + device)
		}
	}

	return true, nil
}
