package wire

import (
	"time"
)

type SetTimeRequest struct {
	CurrentTime time.Time `wire:"type:hexdatetime"` // This is the new time.
	_           [0]byte   `wire:"length:*"`         // Fail if there are any leftover bytes.
}

type SetTimeResponse struct {
	CurrentTime time.Time `wire:"type:hexdatetime"` // This is the new time.
	_           [0]byte   `wire:"length:*"`         // Fail if there are any leftover bytes.
}
