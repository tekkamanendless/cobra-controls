package wire

type DeleteRecordRequest struct {
	RecordIndex uint32  // This is the index to delete; record indexes will be recalculated after this operation.
	Unknown1    []byte  `wire:"length:4"`
	_           [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type DeleteRecordResponse struct {
	Result uint8   // 0 means successs.
	_      [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
