package wire

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type Protocol string

const (
	ProtocolTCP = "tcp"
	ProtocolUDP = "udp"
)

type Client struct {
	Protocol          Protocol
	ControllerAddress string
	ControllerPort    int
	BoardAddress      uint16

	BufferSize int

	conn net.Conn
}

func (c *Client) init() error {
	if c.conn == nil {
		logrus.Debugf("Creating connection.")
		if c.ControllerPort == 0 {
			c.ControllerPort = 60000
		}
		if len(c.Protocol) == 0 {
			c.Protocol = ProtocolTCP
		}
		if c.BufferSize == 0 {
			c.BufferSize = 1024
		}

		var err error
		logrus.Debugf("Dialing (%s): %s:%d", c.Protocol, c.ControllerAddress, c.ControllerPort)
		c.conn, err = net.Dial(string(c.Protocol), fmt.Sprintf("%s:%d", c.ControllerAddress, c.ControllerPort))
		if err != nil {
			return err
		}
		logrus.Debugf("Connected.")
	}
	return nil
}

func (c *Client) Raw(f uint16, request any, response any) error {
	if err := c.init(); err != nil {
		return err
	}

	{
		c.conn.SetDeadline(time.Now().Add(5 * time.Second))

		payloadWriter := NewWriter()
		if request != nil {
			err := Encode(payloadWriter, request)
			if err != nil {
				return fmt.Errorf("could not encode contents: %w", err)
			}
		}

		envelope := Envelope{
			BoardAddress: c.BoardAddress,
			Function:     f,
			Contents:     payloadWriter.Bytes(),
		}

		messageWriter := NewWriter()
		err := Encode(messageWriter, &envelope)
		if err != nil {
			return fmt.Errorf("could not encode envelope: %v", err)
		}
		bytesWritten, err := c.conn.Write(messageWriter.Bytes())
		if err != nil {
			c.conn.Close()
			c.conn = nil
			return fmt.Errorf("could not write message: %v", err)
		}
		logrus.Debugf("Bytes written: %d", bytesWritten)
		if bytesWritten != messageWriter.Length() {
			return fmt.Errorf("could not write full message; wrote %d bytes (expected: %d)", bytesWritten, messageWriter.Length())
		}
	}

	{
		c.conn.SetDeadline(time.Now().Add(5 * time.Second))

		contents := make([]byte, c.BufferSize)
		bytesRead, err := c.conn.Read(contents)
		if err != nil {
			c.conn.Close()
			c.conn = nil
			return fmt.Errorf("could not read contents: %v", err)
		}
		contents = contents[0:bytesRead]
		logrus.Debugf("Bytes read: (%d) %x", bytesRead, contents)

		reader := NewReader(contents)

		var envelope Envelope
		err = Decode(reader, &envelope)
		if err != nil {
			return fmt.Errorf("could not decode envelope: %v", err)
		}
		logrus.Debugf("Response: %x", envelope.Contents)

		if response != nil {
			err = Decode(NewReader(envelope.Contents), response)
			if err != nil {
				return fmt.Errorf("could not decode response: %v", err)
			}
		}
	}

	return nil
}
