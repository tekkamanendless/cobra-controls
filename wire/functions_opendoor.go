package wire

type OpenDoorRequest struct {
	Door     uint8
	Unkonwn1 uint8
	_        [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type OpenDoorResponse struct {
	_ [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
