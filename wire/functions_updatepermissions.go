package wire

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type UpdatePermissionsRequest struct {
	Unknown1  uint16
	CardID    uint16
	Area      uint8
	Door      uint8
	StartDate time.Time
	EndDate   time.Time
	Time      uint8  // TODO: WHAT IS THIS
	Password  uint32 // 24-bit password
	Standby   []byte
}

func (r *UpdatePermissionsRequest) Encode() ([]byte, error) {
	writer := NewWriter()
	writer.WriteUint16(r.Unknown1)
	writer.WriteUint16(r.CardID)
	writer.WriteUint8(r.Area)
	writer.WriteUint8(r.Door)
	writer.WriteDate(r.StartDate)
	writer.WriteDate(r.EndDate)
	writer.WriteUint8(r.Time)
	writer.WriteUint24(r.Password)
	if len(r.Standby) != 4 {
		return nil, fmt.Errorf("not enough bytes for standby: %d (expected: 4)", len(r.Standby))
	}
	writer.WriteBytes(r.Standby)
	return writer.Bytes(), nil
}

func (r *UpdatePermissionsRequest) Decode(b []byte) error {
	var err error
	reader := NewReader(b)
	r.Unknown1, err = reader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read unknown1: %w", err)
	}
	r.CardID, err = reader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read card ID: %w", err)
	}
	r.Area, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read area: %w", err)
	}
	r.Door, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read door: %w", err)
	}
	r.StartDate, err = reader.ReadDate()
	if err != nil {
		return fmt.Errorf("could not read start date: %w", err)
	}
	r.EndDate, err = reader.ReadDate()
	if err != nil {
		return fmt.Errorf("could not read end date: %w", err)
	}
	r.Time, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read time: %w", err)
	}
	r.Password, err = reader.ReadUint24()
	if err != nil {
		return fmt.Errorf("could not read password: %w", err)
	}
	r.Standby, err = reader.ReadBytes(4)
	if err != nil {
		return fmt.Errorf("could not read standby: %w", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", b)
	}
	return nil
}

type UpdatePermissionsResponse struct {
	Result uint8
}

func (r *UpdatePermissionsResponse) Encode() ([]byte, error) {
	writer := NewWriter()
	writer.WriteUint8(r.Result)
	return writer.Bytes(), nil
}

func (r *UpdatePermissionsResponse) Decode(b []byte) error {
	var err error
	reader := NewReader(b)
	r.Result, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read result: %w", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		logrus.Warnf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}
