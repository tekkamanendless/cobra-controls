package wire

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type OpenDoorRequest struct {
	Door     uint8
	Unkonwn1 uint8
}

func (r OpenDoorRequest) Encode() ([]byte, error) {
	writer := NewWriter()
	writer.WriteUint8(r.Door)
	writer.WriteUint8(r.Unkonwn1)
	return writer.Bytes(), nil
}

func (r *OpenDoorRequest) Decode(b []byte) error {
	var err error
	reader := NewReader(b)
	r.Door, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read door: %w", err)
	}
	r.Unkonwn1, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read unknown1: %w", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", b)
	}
	return nil
}

type OpenDoorResponse struct {
}

func (r OpenDoorResponse) Encode() ([]byte, error) {
	writer := NewWriter()
	return writer.Bytes(), nil
}

func (r *OpenDoorResponse) Decode(b []byte) error {
	reader := NewReader(b)
	if !IsAll(reader.Bytes(), 0) {
		logrus.Warnf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}
