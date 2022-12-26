package wire

// RawMessage is a byte array that will be encoded and decoded as such.
type RawMessage []byte

func (r *RawMessage) Encode() ([]byte, error) {
	if len(*r) == 0 {
		return []byte{}, nil
	}
	output := make([]byte, len(*r))
	copy(output, *r)
	return output, nil
}

func (r *RawMessage) Decode(b []byte) error {
	if len(b) == 0 {
		*r = []byte{}
		return nil
	}
	*r = make([]byte, len(b))
	copy(*r, b)
	return nil
}
