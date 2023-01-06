package wire

type GetSettingRequest struct {
	Address  uint8
	Unknown1 uint8
	_        [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type GetSettingResponse struct {
	Value    uint8
	Unknown1 []byte  `wire:"length:*"`
	_        [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
