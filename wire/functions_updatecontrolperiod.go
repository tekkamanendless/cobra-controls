package wire

import "time"

type UpdateControlPeriodRequest struct {
	TimeIndex         uint16
	WeekControl       uint8
	NextLinkTimeIndex uint8
	Standby1          uint8
	Standby2          uint8
	StartTime1        time.Time `wire:"type:time"`
	EndTime1          time.Time `wire:"type:time"`
	StartTime2        time.Time `wire:"type:time"`
	EndTime2          time.Time `wire:"type:time"`
	StartTime3        time.Time `wire:"type:time"`
	EndTime3          time.Time `wire:"type:time"`
	StartDate         time.Time `wire:"type:date"`
	EndDate           time.Time `wire:"type:date"`
	Standby3          uint8
	Standby4          uint8
	Standby5          uint8
	Standby6          uint8
	_                 [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type UpdateControlPeriodResponse UpdateControlPeriodRequest
