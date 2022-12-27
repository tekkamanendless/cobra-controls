package wire

import "time"

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
