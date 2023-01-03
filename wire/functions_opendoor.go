package wire

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type OpenDoorRequest struct {
	Door     uint8
	Unkonwn1 uint8
}

func (r OpenDoorRequest) Encode(writer *Writer) error {
	writer.WriteUint8(r.Door)
	writer.WriteUint8(r.Unkonwn1)
	return nil
}

func (r *OpenDoorRequest) Decode(reader *Reader) error {
	var err error
	r.Door, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read door: %w", err)
	}
	r.Unkonwn1, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read unknown1: %w", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}

type OpenDoorResponse struct {
}

func (r OpenDoorResponse) Encode(writer *Writer) error {
	return nil
}

func (r *OpenDoorResponse) Decode(reader *Reader) error {
	if !IsAll(reader.Bytes(), 0) {
		logrus.Warnf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}
