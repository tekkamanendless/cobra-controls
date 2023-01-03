package wire

import (
	"fmt"
	"time"
)

type GetOperationStatusRequest struct {
	RecordIndex uint32 // 0x0 and 0xFFFFFFFF mean "latest".
}

func (r GetOperationStatusRequest) Encode(writer *Writer) error {
	writer.WriteUint32(r.RecordIndex)
	return nil
}

func (r *GetOperationStatusRequest) Decode(reader *Reader) error {
	var err error
	r.RecordIndex, err = reader.ReadUint32()
	if err != nil {
		return fmt.Errorf("could not read record index: %v", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}
	return nil
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
	BrushDateTime time.Time // This is the time of the access.
}

// Door returns the door with a one index (1-4).
// A value of 0 means invalid door.
func (r *Record) Door() uint8 {
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

func (r *Record) AccessGranted() bool {
	return (r.RecordState&0b10000000 == 0)
}

func (r Record) Encode(writer *Writer) error {
	writer.WriteUint16(r.IDNumber)
	writer.WriteUint8(r.AreaNumber)
	writer.WriteUint8(r.RecordState)
	writer.WriteDate(r.BrushDateTime)
	writer.WriteTime(r.BrushDateTime)
	return nil
}

func (r *Record) Decode(reader *Reader) error {
	var err error
	r.IDNumber, err = reader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read ID number: %v", err)
	}
	r.AreaNumber, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read area number: %v", err)
	}
	r.RecordState, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read record state: %v", err)
	}
	brushDate, err := reader.ReadDate()
	if err != nil {
		return fmt.Errorf("could not read brush date: %v", err)
	}
	brushTime, err := reader.ReadTime()
	if err != nil {
		return fmt.Errorf("could not read brush time: %v", err)
	}
	r.BrushDateTime = MergeDateTime(brushDate, brushTime)
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}

type GetOperationStatusResponse struct {
	CurrentTime   time.Time // TODO: This is supposed to be the time, but the format makes no sense.
	RecordCount   uint32    // This is the number of access records available.
	PopedomAmount uint16    // TODO: Is the number of fobs registered on the door?
	Record        *Record   // This is the access record for the index requested.
	RelayStatus   uint8
	MagnetState   uint8
	Reserved1     uint8
	FaultNumber   uint8
	Reserved2     uint8
	Reserved3     uint8
}

func (r GetOperationStatusResponse) Encode(writer *Writer) error {
	year := r.CurrentTime.Year()
	if year > 2000 {
		year -= 2000
	}
	writer.WriteUint8(InsaneBase10ToBase16(uint8(year)))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.CurrentTime.Month())))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.CurrentTime.Day())))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.CurrentTime.Weekday())))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.CurrentTime.Hour())))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.CurrentTime.Minute())))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.CurrentTime.Second())))
	writer.WriteUint24(r.RecordCount)
	writer.WriteUint16(r.PopedomAmount)
	if r.Record == nil {
		recordBytes := make([]byte, 8)
		for b := range recordBytes {
			recordBytes[b] = 0xff
		}
		writer.WriteBytes(recordBytes)
	} else {
		err := r.Record.Encode(writer)
		if err != nil {
			return fmt.Errorf("could not encode record: %w", err)
		}
	}
	writer.WriteUint8(r.RelayStatus)
	writer.WriteUint8(r.MagnetState)
	writer.WriteUint8(r.Reserved1)
	writer.WriteUint8(r.FaultNumber)
	writer.WriteUint8(r.Reserved2)
	writer.WriteUint8(r.Reserved3)
	return nil
}

func (r *GetOperationStatusResponse) Decode(reader *Reader) error {
	var err error
	year, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read year: %w", err)
	}
	year = InsaneBase16ToBase10(year)
	month, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read month: %w", err)
	}
	month = InsaneBase16ToBase10(month)
	day, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read day: %w", err)
	}
	day = InsaneBase16ToBase10(day)
	week, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read week: %w", err)
	}
	week = InsaneBase16ToBase10(week)
	hour, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read hour: %w", err)
	}
	hour = InsaneBase16ToBase10(hour)
	minute, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read minute: %w", err)
	}
	minute = InsaneBase16ToBase10(minute)
	second, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read second: %w", err)
	}
	second = InsaneBase16ToBase10(second)
	r.CurrentTime = time.Date(2000+int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)
	_ = week

	r.RecordCount, err = reader.ReadUint24()
	if err != nil {
		return fmt.Errorf("could not read card record: %w", err)
	}
	r.PopedomAmount, err = reader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read popedom amount: %w", err)
	}
	recordReader, err := reader.Read(8)
	if err != nil {
		return fmt.Errorf("could not read record: %w", err)
	}
	if !IsAll(recordReader.Bytes(), 0x00) && !IsAll(recordReader.Bytes(), 0xff) {
		var record Record
		err = Decode(recordReader, &record)
		if err != nil {
			return fmt.Errorf("could not parse record: %w", err)
		}
		r.Record = &record
	}
	r.RelayStatus, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read relay status: %w", err)
	}
	r.MagnetState, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read door magnet button state: %w", err)
	}
	r.Reserved1, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read reserved 1: %w", err)
	}
	r.FaultNumber, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read reserved 2: %w", err)
	}
	r.Reserved2, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read reserved 3: %w", err)
	}
	r.Reserved3, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read reserved 4: %w", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}

	return nil
}
