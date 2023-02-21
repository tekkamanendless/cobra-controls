package wire

import (
	"time"
)

type DeletePermissionsRequest struct {
	Empty1    uint16 // This should always be zero.
	CardID    uint16
	Area      uint8
	Door      uint8
	StartDate *time.Time `wire:"type:date,null:0x00"`
	EndDate   *time.Time `wire:"type:date,null:0x00"`
	Time      uint8      // TODO: WHAT IS THIS
	Password  uint32     `wire:"type:uint24"` // 24-bit password
	Standby   []byte     `wire:"length:4"`
	_         [0]byte    `wire:"length:*"` // Fail if there are any leftover bytes.
}

type DeletePermissionsResponse struct {
	Result uint8
	_      [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
