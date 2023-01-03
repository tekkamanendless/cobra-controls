package wire

// RawMessage is a byte array that will be encoded and decoded as such.
type RawMessage []byte

func (r RawMessage) Encode(writer *Writer) error {
	writer.WriteBytes(r)
	return nil
}

func (r *RawMessage) Decode(reader *Reader) error {
	b, err := reader.ReadBytes(reader.Length())
	if err != nil {
		return err
	}
	*r = make([]byte, len(b))
	copy(*r, b)
	return nil
}
