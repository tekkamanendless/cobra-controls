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
	Unknown1  []byte  `wire:"length:21"`
	_         [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
