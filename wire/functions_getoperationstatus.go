package wire

import (
	"time"
)

type GetOperationStatusRequest struct {
	RecordIndex uint32  // 0x0 and 0xFFFFFFFF mean "latest".
	_           [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

// TODO: RecordState
//
// Card ID seems to refer to "${AreaNumber}${IDNumber}".
// It doesn't make any sense why you'd compare this to 100; at first I thought
// that it only referred to the area number, but I checked a bunch of our IDs,
// and there are some with a low area number.  I suspect that this only means
// that (1) area number is zero and (2) ID number is less than 100.
//
// When the card ID is over 100...
// Relay state bits:
// 876543 21
// State  Door
// (xx represents the door, 0-3.)
// 100000 xx Denied, "non-specific"
// 100100 xx Denied, "do not have permission"
// 101000 xx Denied, "password incorrect"
// 101100 xx Denied, "system at fault"
// 110000 xx Denied, "anti-submarine back, many cards to open the door or door interlocking"
// 110001 xx Denied, "anti-submarine back"
// 110010 xx Denied, "many cards"
// 110011 xx Denied, "the first card"
// 110100 xx Denied, "the door is normal closed"
// 110101 xx Denied, "door interlocking"
// 111000 xx Denied, "card expired or not valid time"
//
// When the card ID is under 100 (it's a special record)...
// Empirically, it would seem that the "area number" is 0 in this case.
// Card bits | Relay state bits:
// 43   21     876543 21
// (xx represents the door, 0-3.)
// Card   Relay state
// 00 xx  000000 00 "button"
// 00 xx  000000 11 "long-distance open" (for example, area=0, card-id=2, record-state=3 means that door 3 was opened by remote control)
// 01 01  000000 xx "super password open"
// 10 xx  000000 00 "door opening, magnetism signal"
// 11 xx  000000 00 "door closed, magnetism signal"
// 00 xx  100000 01 "duress alarm"
// 00 xx  100000 10 "long time not close alarm"
// 00 xx  100001 00 "illegal intrusion alarm"
// 01 00  101000 00 "fire alarm action (whole controller)"
// 01 10  101000 00 "compulsion lock door (whole controller)"

type Record struct {
	IDNumber      uint16    // "${AreaNumber}${IDNumber}" is the fob ID in the UI.
	AreaNumber    uint8     // "${AreaNumber}${IDNumber}" is the fob ID in the UI.
	RecordState   uint8     // This is the state that has been recorded (access granted/denied, which door, etc.).
	BrushDateTime time.Time `wire:"type:datetime"` // This is the time of the access.
	_             [0]byte   `wire:"length:*"`      // Fail if there are any leftover bytes.
}

// Door returns the door with a one index (1-4).
// A value of 0 means invalid door.
func (r Record) Door() uint8 {
	if r.AreaNumber == 0 && r.IDNumber < 100 {
		if r.IDNumber&0b1100 == 0b0000 {
			return (uint8(r.IDNumber) & 0b11) + 1
		}
		if r.IDNumber&0b1100 == 0b1000 {
			return (uint8(r.IDNumber) & 0b11) + 1
		}
		if r.IDNumber&0b1100 == 0b1100 {
			return (uint8(r.IDNumber) & 0b11) + 1
		}
		if r.IDNumber == 0b0101 {
			return (r.RecordState & 0b11) + 1
		}
		if r.IDNumber == 0b0100 {
			return 0 // No specific door.
		}
		if r.IDNumber == 0b0110 {
			return 0 // No specific door.
		}
	}
	return (r.RecordState & 0b11) + 1
}

func (r Record) AccessGranted() bool {
	return (r.RecordState&0b10000000 == 0)
}

type GetOperationStatusResponse struct {
	CurrentTime   time.Time `wire:"type:hexdatetime"`
	RecordCount   uint32    `wire:"type:uint24"` // This is the number of access records available.
	PopedomAmount uint16    // TODO: Is the number of fobs registered on the door?
	Record        *Record   `wire:"length:8,null:0xff"` // This is the access record for the index requested.
	RelayStatus   uint8
	MagnetState   uint8
	Reserved1     uint8
	FaultNumber   uint8
	Reserved2     uint8
	Reserved3     uint8
	_             [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
