package wire

import (
	"time"
)

// IsAll returns true if all bytes in the slice are the specified value.
//
// If the slice is empty, then this returns true.
func IsAll(data []byte, expectedValue byte) bool {
	for _, b := range data {
		if b != expectedValue {
			return false
		}
	}
	return true
}

// MergeDateTime takes the date portion from `dateOnly` and the time portion
// from `timeOnly` and returns a single timestamp.
func MergeDateTime(dateOnly, timeOnly time.Time) time.Time {
	return time.Date(dateOnly.Year(), dateOnly.Month(), dateOnly.Day(), timeOnly.Hour(), timeOnly.Minute(), timeOnly.Second(), 0, time.UTC)
}

// InsaneBase16ToBase10 is an insane decoder function.
// The value is the base-10-printable version of a printed hexadecimal value.
//
// For example, 34 is 0x22, so the result is 22.
func InsaneBase16ToBase10(value uint8) uint8 {
	var result uint8
	for i := uint8(1); value > 0; i *= 10 {
		remainder := value % 16
		result += remainder * i
		value /= 16
	}
	return result
}

// InsaneBase10ToBase16 is an insane encoder function.
// The value is the base-16 rendering of a base-10 value.
//
// For example, 0x22 is is 34, so the value is 34.
func InsaneBase10ToBase16(value uint8) uint8 {
	var result uint8
	for i := uint8(1); value > 0; i *= 16 {
		remainder := value % 10
		result += remainder * i
		value /= 10
	}
	return result
}
