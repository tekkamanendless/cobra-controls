package wire

import (
	"fmt"
)

const (
	EnvelopeStartByte = 0x7E
	EnvelopeEndByte   = 0x0D
)

type Envelope struct {
	BoardAddress uint16 // This is the board address.  It appears to be the last 2 bytes of the MAC address.
	Function     uint16 // This is the function.
	Contents     []byte // This is the contents of the message.
}

func (e *Envelope) Encode() ([]byte, error) {
	w := NewWriter()
	w.WriteUint16(e.BoardAddress)
	w.WriteUint16(e.Function)
	w.WriteBytes(e.Contents)

	internalContents := w.Bytes()
	checksum := uint16(0)
	for i := 0; i < len(internalContents); i++ {
		checksum += uint16(internalContents[i])
	}

	w = NewWriter()
	w.WriteUint8(EnvelopeStartByte)
	w.WriteBytes(internalContents)
	w.WriteUint16(checksum)
	w.WriteUint8(EnvelopeEndByte)

	return w.Bytes(), nil
}

func (e *Envelope) Decode(contents []byte) error {
	r := NewReader(contents)

	startByte, err := r.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read start byte: %w", err)
	}

	internalContents, err := r.ReadBytes(r.Length() - 3)
	if err != nil {
		return fmt.Errorf("could not read internal contents: %w", err)
	}

	expectedChecksum, err := r.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read checksum: %w", err)
	}
	endByte, err := r.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read end byte: %w", err)
	}

	if startByte != EnvelopeStartByte {
		return fmt.Errorf("invalid start byte: 0x%x (expected: 0x%x)", startByte, EnvelopeStartByte)
	}
	if endByte != EnvelopeEndByte {
		return fmt.Errorf("invalid end byte: 0x%x (expected: 0x%x)", endByte, EnvelopeEndByte)
	}
	actualChecksum := uint16(0)
	for i := 0; i < len(internalContents); i++ {
		actualChecksum += uint16(internalContents[i])
	}
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("invalid checksum: %d (expected: %d)", actualChecksum, expectedChecksum)
	}

	r = NewReader(internalContents)
	e.BoardAddress, err = r.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read board address: %w", err)
	}
	e.Function, err = r.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read function: %w", err)
	}
	e.Contents, err = r.ReadBytes(r.Length() - 3)
	if err != nil {
		return fmt.Errorf("could not read contents: %w", err)
	}
	return nil
}
