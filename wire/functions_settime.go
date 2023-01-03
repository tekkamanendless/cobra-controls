package wire

import (
	"fmt"
	"time"
)

type SetTimeRequest struct {
	CurrentTime time.Time // This is the new time.
	_           [0]byte   `wire:"length:*"` // Fail if there are any leftover bytes.
}

func (r SetTimeRequest) Encode(writer *Writer) error {
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
	return nil
}

func (r *SetTimeRequest) Decode(reader *Reader) error {
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

	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}

	return nil
}

type SetTimeResponse struct {
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
	_             [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

func (r SetTimeResponse) Encode(writer *Writer) error {
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
	return nil
}

func (r *SetTimeResponse) Decode(reader *Reader) error {
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

	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}

	return nil
}
