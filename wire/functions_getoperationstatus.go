package wire

import (
	"fmt"
	"time"
)

type GetOperationStatusRequest struct {
	RecordIndex uint32 // 0x0 and 0xFFFFFFFF mean "latest".
}

func (r *GetOperationStatusRequest) Encode() ([]byte, error) {
	writer := NewWriter()
	writer.WriteUint32(r.RecordIndex)
	return writer.Bytes(), nil
}

func (r *GetOperationStatusRequest) Decode(b []byte) error {
	reader := NewReader(b)
	var err error
	r.RecordIndex, err = reader.ReadUint32()
	if err != nil {
		return fmt.Errorf("could not read record index: %v", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", b)
	}
	return nil
}

type Record struct {
	IDNumber      uint16    // "${AreaNumber}${IDNumber}" is the fob ID in the UI.
	AreaNumber    uint8     // "${AreaNumber}${IDNumber}" is the fob ID in the UI.
	RecordState   uint8     // TODO: This appears to be the door.
	BrushDateTime time.Time // This is the time of the access.
}

func (r *Record) Encode() ([]byte, error) {
	writer := NewWriter()
	writer.WriteUint16(r.IDNumber)
	writer.WriteUint8(r.AreaNumber)
	writer.WriteUint8(r.RecordState)
	writer.WriteDate(r.BrushDateTime)
	writer.WriteTime(r.BrushDateTime)
	return writer.Bytes(), nil
}

func (r *Record) Decode(b []byte) error {
	reader := NewReader(b)
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
		return fmt.Errorf("unexpected contents: %x", b)
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

func (r *GetOperationStatusResponse) Encode() ([]byte, error) {
	writer := NewWriter()
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
		recordBytes, err := r.Record.Encode()
		if err != nil {
			return nil, fmt.Errorf("could not encode record: %w", err)
		}
		writer.WriteBytes(recordBytes)
	}
	writer.WriteUint8(r.RelayStatus)
	writer.WriteUint8(r.MagnetState)
	writer.WriteUint8(r.Reserved1)
	writer.WriteUint8(r.FaultNumber)
	writer.WriteUint8(r.Reserved2)
	writer.WriteUint8(r.Reserved3)
	return writer.Bytes(), nil
}

func (r *GetOperationStatusResponse) Decode(b []byte) error {
	reader := NewReader(b)
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
	recordBytes, err := reader.ReadBytes(8)
	if err != nil {
		return fmt.Errorf("could not read record: %w", err)
	}
	if !IsAll(recordBytes, 0x00) && !IsAll(recordBytes, 0xff) {
		var record Record
		err = Decode(recordBytes, &record)
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
		return fmt.Errorf("unexpected contents: %x", b)
	}

	return nil
}