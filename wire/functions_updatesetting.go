package wire

type UpdateSettingRequest struct {
	Address  uint8
	Unknown1 uint8
	Value    uint8
	_        [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type UpdateSettingResponse struct {
	Result uint8   // 0 means successs.
	_      [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
