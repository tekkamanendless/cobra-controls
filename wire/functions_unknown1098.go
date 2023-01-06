package wire

type Unknown1098Request struct {
	_ [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type Unknown1098Response struct {
	Result uint8   // 0 means successs.
	_      [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
