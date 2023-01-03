package wire

import "time"

type GetRecordRequest struct {
	RecordIndex uint32
	_           [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type GetRecordResponse struct {
	CardNumber        uint16
	AreaNumber        uint8
	BrushCardState    uint8
	BrushCardDateTime time.Time `wire:"type:datetime"`
	Unknown1          []byte    `wire:"length:*"`
	_                 [0]byte   `wire:"length:*"` // Fail if there are any leftover bytes.
}
