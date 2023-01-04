package wire

import (
	"time"
)

type GetBasicInfoRequest struct {
	_ [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type GetBasicInfoResponse struct {
	IssueDate time.Time `wire:"type:hexdate"`
	Version   uint8
	Model     uint8
	Unknown2  []byte  `wire:"length:*"`
	_         [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
