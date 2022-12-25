package wire

// Encoder can encode an object.
type Encoder interface {
	Encode() ([]byte, error)
}

// Decoder can decode an object.
type Decoder interface {
	Decode([]byte) error
}

// Encode an object to the wire.
func Encode(e Encoder) ([]byte, error) {
	return e.Encode()
}

// Decode an object from the wire.
func Decode(data []byte, d Decoder) error {
	return d.Decode(data)
}
