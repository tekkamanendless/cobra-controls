package wire

import (
	"fmt"
	"time"
)

type Reader struct {
	data []byte
}

func NewReader(data []byte) *Reader {
	r := &Reader{
		data: []byte{},
	}
	if len(data) > 0 {
		r.data = make([]byte, len(data))
		copy(r.data, data)
	}
	return r
}

func (r *Reader) Bytes() []byte {
	return r.data
}

func (r *Reader) Length() int {
	return len(r.data)
}

func (r *Reader) ReadBytes(count int) ([]byte, error) {
	if count == 0 {
		return []byte{}, nil
	}
	if len(r.data) < count {
		return nil, fmt.Errorf("data too small; only have %d bytes available (asked for: %d)", len(r.data), count)
	}
	output := r.data[0:count]
	r.data = r.data[count:]
	return output, nil
}

func (r *Reader) Read(count int) (*Reader, error) {
	contents, err := r.ReadBytes(count)
	if err != nil {
		return nil, err
	}
	return NewReader(contents), nil
}

func (r *Reader) ReadUint8() (uint8, error) {
	contents, err := r.ReadBytes(1)
	if err != nil {
		return 0, err
	}
	return contents[0], nil
}

func (r *Reader) ReadUint16() (uint16, error) {
	contents, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return (uint16(contents[1]) << 8) | uint16(contents[0]), nil
}

func (r *Reader) ReadUint24() (uint32, error) {
	contents, err := r.ReadBytes(3)
	if err != nil {
		return 0, err
	}
	return (uint32(contents[2]) << 16) | (uint32(contents[1]) << 8) | uint32(contents[0]), nil
}

func (r *Reader) ReadUint32() (uint32, error) {
	contents, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return (uint32(contents[3]) << 24) | (uint32(contents[2]) << 16) | (uint32(contents[1]) << 8) | uint32(contents[0]), nil
}

func (r *Reader) ReadDate() (time.Time, error) {
	value, err := r.ReadUint16()
	if err != nil {
		return time.Time{}, err
	}
	year := (value & 0b1111111000000000) >> 9
	month := (value & 0b0000000111100000) >> 5
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("invalid month: %d", month)
	}
	day := (value & 0b0000000000011111) >> 0
	if month < 1 || month > 31 {
		return time.Time{}, fmt.Errorf("invalid day: %d", month)
	}

	output := time.Date(2000+int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	return output, nil
}

func (r *Reader) ReadTime() (time.Time, error) {
	value, err := r.ReadUint16()
	if err != nil {
		return time.Time{}, err
	}
	hours := (value & 0b1111100000000000) >> 11
	minutes := (value & 0b0000011111100000) >> 5
	seconds := (value & 0b0000000000011111) >> 0

	if hours >= 24 {
		return time.Time{}, fmt.Errorf("invalid hour: %d", hours)
	}
	if minutes >= 60 {
		return time.Time{}, fmt.Errorf("invalid minute: %d", minutes)
	}
	if seconds >= 60 {
		return time.Time{}, fmt.Errorf("invalid second: %d", seconds)
	}

	output := time.Date(0, time.January, 1, int(hours), int(minutes), int(seconds)*2, 0, time.UTC)
	return output, nil
}
