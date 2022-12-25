package wire

import (
	"bytes"
	"time"
)

type Writer struct {
	buffer bytes.Buffer
}

func NewWriter() *Writer {
	w := &Writer{
		buffer: bytes.Buffer{},
	}
	return w
}

func (w *Writer) Bytes() []byte {
	return w.buffer.Bytes()
}

func (w *Writer) Length() int {
	return w.buffer.Len()
}

func (w *Writer) WriteBytes(b []byte) {
	w.buffer.Write(b)
}

func (w *Writer) Write(writer *Writer) {
	w.WriteBytes(writer.Bytes())
}

func (w *Writer) WriteUint8(value uint8) {
	w.WriteBytes([]byte{value})
}

func (w *Writer) WriteUint16(value uint16) {
	w.WriteBytes([]byte{
		byte(value & 0x00ff),
		byte((value & 0xff00) >> 8),
	})
}

func (w *Writer) WriteUint24(value uint32) {
	w.WriteBytes([]byte{
		byte(value & 0x0000ff),
		byte((value & 0x00ff00) >> 8),
		byte((value & 0xff0000) >> 16),
	})
}

func (w *Writer) WriteUint32(value uint32) {
	w.WriteBytes([]byte{
		byte(value & 0x000000ff),
		byte((value & 0x0000ff00) >> 8),
		byte((value & 0x00ff0000) >> 16),
		byte((value & 0xff000000) >> 24),
	})
}

func (w *Writer) WriteDate(value time.Time) {
	year := uint16(value.Year())
	if year >= 2000 {
		year -= 2000
	}
	month := uint16(value.Month())
	day := uint16(value.Day())

	output := ((year & 0b1111111) << 9) | ((month & 0b1111) << 5) | ((day & 0b11111) << 0)
	w.WriteUint16(output)
}

func (w *Writer) WriteTime(value time.Time) {
	hours := uint16(value.Hour())
	minutes := uint16(value.Minute())
	seconds := uint16(value.Second() / 2)

	output := ((hours & 0b11111) << 11) | ((minutes & 0b111111) << 5) | ((seconds & 0b11111) << 0)
	w.WriteUint16(output)
}
