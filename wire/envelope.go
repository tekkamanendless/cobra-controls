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

func (e *Envelope) Encode(writer *Writer) error {
	internalWriter := NewWriter()
	internalWriter.WriteUint16(e.BoardAddress)
	internalWriter.WriteUint16(e.Function)
	internalWriter.WriteBytes(e.Contents)
	// Pad the contents to 26 bytes.  More than 26 bytes is fine.
	for i := 0; i < 26-len(e.Contents); i++ {
		internalWriter.WriteUint8(0)
	}

	internalContents := internalWriter.Bytes()
	checksum := uint16(0)
	for i := 0; i < len(internalContents); i++ {
		checksum += uint16(internalContents[i])
	}

	writer.WriteUint8(EnvelopeStartByte)
	writer.WriteBytes(internalContents)
	writer.WriteUint16(checksum)
	writer.WriteUint8(EnvelopeEndByte)
	return nil
}

func (e *Envelope) Decode(reader *Reader) error {
	startByte, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read start byte: %w", err)
	}
	if startByte != EnvelopeStartByte {
		return fmt.Errorf("invalid start byte: 0x%x (expected: 0x%x)", startByte, EnvelopeStartByte)
	}

	internalContents, err := reader.ReadBytes(reader.Length() - 3) // 2 bytes for the checksum and 1 for the end byte.
	if err != nil {
		return fmt.Errorf("could not read internal contents: %w", err)
	}

	if reader.Length() != 3 {
		return fmt.Errorf("somehow did not read enough data; length is %d (expected: %d)", reader.Length(), 3)
	}

	expectedChecksum, err := reader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read checksum: %w", err)
	}
	endByte, err := reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read end byte: %w", err)
	}
	if endByte != EnvelopeEndByte {
		return fmt.Errorf("invalid end byte: 0x%x (expected: 0x%x)", endByte, EnvelopeEndByte)
	}

	if reader.Length() != 0 {
		return fmt.Errorf("somehow did not read enough data; length is %d", reader.Length())
	}

	actualChecksum := uint16(0)
	for i := 0; i < len(internalContents); i++ {
		actualChecksum += uint16(internalContents[i])
	}
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("invalid checksum: %d (expected: %d)", actualChecksum, expectedChecksum)
	}

	payloadReader := NewReader(internalContents)
	e.BoardAddress, err = payloadReader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read board address: %w", err)
	}
	e.Function, err = payloadReader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read function: %w", err)
	}
	e.Contents, err = payloadReader.ReadBytes(payloadReader.Length())
	if err != nil {
		return fmt.Errorf("could not read contents: %w", err)
	}
	return nil
}
