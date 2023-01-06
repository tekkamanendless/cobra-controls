package wire

type ClearUploadRequest struct {
	_ [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type ClearUploadResponse struct {
	Result uint8   // 0 means successs.
	_      [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
