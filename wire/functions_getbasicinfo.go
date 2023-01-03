package wire

import (
	"fmt"
	"time"
)

type GetBasicInfoRequest struct{}

func (r GetBasicInfoRequest) Encode(writer *Writer) error {
	return nil
}

func (r *GetBasicInfoRequest) Decode(reader *Reader) error {
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}

type GetBasicInfoResponse struct {
	IssueDate time.Time
	Version   uint8
	Model     uint8
	Unknown2  []byte
}

func (r GetBasicInfoResponse) Encode(writer *Writer) error {
	year := r.IssueDate.Year()
	if year > 2000 {
		year -= 2000
	}
	writer.WriteUint8(InsaneBase10ToBase16(uint8(year)))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.IssueDate.Month())))
	writer.WriteUint8(InsaneBase10ToBase16(uint8(r.IssueDate.Day())))
	writer.WriteUint8(r.Version)
	writer.WriteUint8(r.Model)
	writer.WriteBytes(r.Unknown2)
	return nil
}

func (r *GetBasicInfoResponse) Decode(reader *Reader) error {
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
	r.IssueDate = time.Date(2000+int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	r.Version, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read version: %w", err)
	}
	r.Model, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read model: %w", err)
	}
	r.Unknown2, err = reader.ReadBytes(reader.Length())
	if err != nil {
		return fmt.Errorf("could not read unknown2: %w", err)
	}
	return nil
}
