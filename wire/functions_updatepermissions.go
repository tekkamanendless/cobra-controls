package wire

import (
	"time"
)

type UpdatePermissionsRequest struct {
	Unknown1  uint16
	CardID    uint16
	Area      uint8
	Door      uint8
	StartDate time.Time `wire:"type:date"`
	EndDate   time.Time `wire:"type:date"`
	Time      uint8     // TODO: WHAT IS THIS
	Password  uint32    `wire:"type:uint24"` // 24-bit password
	Standby   []byte    `wire:"length:4"`
	_         [0]byte   `wire:"length:*"` // Fail if there are any leftover bytes.
}

type UpdatePermissionsResponse struct {
	Result uint8
	_      [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
