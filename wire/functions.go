package wire

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	FunctionGetBasicInfo = 0x1082
)

type GetBasicInfoRequest struct{}

func (r *GetBasicInfoRequest) Encode() ([]byte, error) {
	return []byte{}, nil
}

func (r *GetBasicInfoRequest) Decode(b []byte) error {
	if !IsAll(b, 0) {
		return fmt.Errorf("unexpected contents: %x", b)
	}
	return nil
}

type GetBasicInfoResponse struct {
	IssueDate time.Time
	Version   uint8
	Model     uint8
}

func (r *GetBasicInfoResponse) Encode() ([]byte, error) {
	writer := NewWriter()
	writer.WriteDate(r.IssueDate)
	writer.WriteUint8(r.Version)
	writer.WriteUint8(r.Model)
	return writer.Bytes(), nil
}

func (r *GetBasicInfoResponse) Decode(b []byte) error {
	var err error
	reader := NewReader(b)
	r.IssueDate, err = reader.ReadDate()
	if err != nil {
		return fmt.Errorf("could not read issue date: %w", err)
	}
	r.Version, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read version: %w", err)
	}
	r.Model, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read model: %w", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		logrus.Warnf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}
