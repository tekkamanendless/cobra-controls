package wire

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
