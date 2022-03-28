package hr_bytes

import "fmt"

type Byte int64

const (
	// B byte
	B = (Byte)(1 << (10 * iota))
	// KB kilobyte
	KB
	// MB megabyte
	MB
	// GB gigabyte
	GB
	// TB terabyte
	TB
	// PB petabyte
	PB
)

func (size Byte) String() string {
	if size < 0 {
		return "0B"
	}
	if size < KB {
		return fmt.Sprintf("%dB", size)
	}
	if size < MB {
		return fmt.Sprintf("%.3fKB", float64(size)/float64(KB))
	}
	if size < GB {
		return fmt.Sprintf("%.3fMB", float64(size)/float64(MB))
	}
	if size < TB {
		return fmt.Sprintf("%.3fGB", float64(size)/float64(GB))
	}
	if size < PB {
		return fmt.Sprintf("%.3fTB", float64(size)/float64(TB))
	}
	return fmt.Sprintf("%.3fPB", float64(size)/float64(PB))
}
