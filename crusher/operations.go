package crusher

import (
	"bytes"
	"math/rand"
	"sort"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// clampBounds ensures start and end are within the bounds of the array.
func clampBounds(length, start, end int) (int, int) {
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		start = end
	}
	return start, end
}

// Corrupt randomizes a percentage of bytes in the data slice between start and end.
func Corrupt(data []byte, start, end int, severity float64) {
	start, end = clampBounds(len(data), start, end)
	if start == end || severity <= 0 {
		return
	}
	if severity > 1.0 {
		severity = 1.0
	}

	workAreaSize := end - start
	bytesToCorrupt := int(float64(workAreaSize) * severity)

	for i := 0; i < bytesToCorrupt; i++ {
		idx := start + rand.Intn(workAreaSize)
		data[idx] = byte(rand.Intn(256))
	}
}

// ReplacePattern replaces instances of 'from' with 'to' sequence within the start and end range.
func ReplacePattern(data []byte, start, end int, from []byte, to []byte) []byte {
	start, end = clampBounds(len(data), start, end)
	if len(from) == 0 || start == end {
		return data
	}

	head := data[:start]
	body := data[start:end]
	tail := data[end:]

	newBody := bytes.ReplaceAll(body, from, to)

	result := make([]byte, 0, len(head)+len(newBody)+len(tail))
	result = append(result, head...)
	result = append(result, newBody...)
	result = append(result, tail...)

	return result
}

// ByteShift shifts the values of bytes up or down within the range.
func ByteShift(data []byte, start, end int, shift int8) {
	start, end = clampBounds(len(data), start, end)
	for i := start; i < end; i++ {
		// using unit8/int8 cast to safely wrap around bounds (0-255)
		data[i] = byte(int16(data[i]) + int16(shift))
	}
}

// Bitwise applies a bitwise mathematical operation using the key to the range.
func Bitwise(data []byte, start, end int, key byte, op string) {
	start, end = clampBounds(len(data), start, end)
	for i := start; i < end; i++ {
		switch op {
		case "xor":
			data[i] = data[i] ^ key
		case "and":
			data[i] = data[i] & key
		case "or":
			data[i] = data[i] | key
		}
	}
}

// SortChunk sorts the bytes in the specified range.
func SortChunk(data []byte, start, end int, ascending bool) {
	start, end = clampBounds(len(data), start, end)
	chunk := data[start:end]
	if ascending {
		sort.Slice(chunk, func(i, j int) bool { return chunk[i] < chunk[j] })
	} else {
		sort.Slice(chunk, func(i, j int) bool { return chunk[i] > chunk[j] })
	}
}

// ReverseChunk reverses the array of bytes in the specified range.
func ReverseChunk(data []byte, start, end int) {
	start, end = clampBounds(len(data), start, end)
	for i, j := start, end-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// PatternOverwrite repeats the given pattern across the specified range, overwriting it.
func PatternOverwrite(data []byte, start, end int, pattern []byte) {
	start, end = clampBounds(len(data), start, end)
	if len(pattern) == 0 || start == end {
		return
	}
	pLen := len(pattern)
	for i := start; i < end; i++ {
		data[i] = pattern[(i-start)%pLen]
	}
}
