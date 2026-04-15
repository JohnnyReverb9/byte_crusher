package crusher

import (
	"bytes"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Corrupt randomizes a percentage of bytes in the data slice, starting after the offset.
func Corrupt(data []byte, offset int, severity float64) {
	if severity <= 0 || len(data) <= offset {
		return
	}
	if severity > 1.0 {
		severity = 1.0
	}

	workAreaSize := len(data) - offset
	bytesToCorrupt := int(float64(workAreaSize) * severity)

	for i := 0; i < bytesToCorrupt; i++ {
		// Pick a random index within the work area
		idx := offset + rand.Intn(workAreaSize)
		// Assign a random byte
		data[idx] = byte(rand.Intn(256))
	}
}

// ReplacePattern replaces all non-overlapping instances of 'from' with 'to' sequence after the given offset.
func ReplacePattern(data []byte, offset int, from []byte, to []byte) []byte {
	if len(from) == 0 || len(data) <= offset {
		return data
	}

	header := data[:offset]
	body := data[offset:]

	// Replace bytes in the body
	newBody := bytes.ReplaceAll(body, from, to)

	// Combine header with the new body
	result := append([]byte{}, header...)
	result = append(result, newBody...)

	return result
}
