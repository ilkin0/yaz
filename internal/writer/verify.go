package writer

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

const verifyBufferSize = 4 * 1024 * 1024

func Verify(device, filePath string, onProgress ProgressFunc) (bool, error) {
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
	var smoothSpeed float64
	lastTime := time.Now()
	var lastVerified uint64

	for {
		fn, ferr := io.ReadFull(file, fileBuf)
		dn, derr := io.ReadFull(d, deviceBuf[:fn]) // read same amount from device

		if fn > 0 {
			if dn != fn {
				return false, errors.New("device read returned fewer bytes than image")
			}

			if !bytes.Equal(fileBuf[:fn], deviceBuf[:fn]) {
				return false, nil
			}

			verified += uint64(fn)

			if onProgress != nil {
				now := time.Now()
				elapsed := now.Sub(lastTime).Seconds()
				if elapsed > 0.5 {
					chunkSpeed := float64(verified-lastVerified) / elapsed
					if smoothSpeed == 0 {
						smoothSpeed = chunkSpeed
					} else {
						smoothSpeed = 0.3*chunkSpeed + 0.7*smoothSpeed
					}
					lastTime = now
					lastVerified = verified
				}

				onProgress(Progress{
					Phase:        PhaseVerifying,
					BytesWritten: verified,
					TotalBytes:   totalBytes,
					Speed:        smoothSpeed,
				})
			}
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
