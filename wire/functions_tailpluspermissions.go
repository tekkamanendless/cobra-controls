package wire

import "time"

type TailPlusPermissionsRequest struct {
	UploadIndex uint16
	CardNumber  uint16
	AreaNumber  uint8
	Door        uint8
	StartDate   time.Time `wire:"type:date"`
	EndDate     time.Time `wire:"type:date"`
	Time        uint8
	Password    uint32 `wire:"type:uint24"`
	Standby1    uint8
	Standby2    uint8
	Standby3    uint8
	Standby4    uint8
	_           [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type TailPlusPermissionsResponse struct {
	Result uint8   // 1 is success; 0 is failure.
	_      [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
