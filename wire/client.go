package wire

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type Client struct {
	ControllerAddress string
	ControllerPort    int
	BoardAddress      uint16

	conn net.Conn
}

func (c *Client) init() error {
	if c.conn == nil {
		logrus.Debugf("Creating connection.")
		if c.ControllerPort == 0 {
			c.ControllerPort = 60000
		}

		var err error
		logrus.Debugf("Dialing: %s:%d", c.ControllerAddress, c.ControllerPort)
		c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", c.ControllerAddress, c.ControllerPort))
		if err != nil {
			return err
		}
		logrus.Debugf("Connected.")
	}
	return nil
}

func (c *Client) Raw(f uint16, e Encoder, d Decoder) error {
	if err := c.init(); err != nil {
		return err
	}

	{
		c.conn.SetDeadline(time.Now().Add(5 * time.Second))

		var contents []byte
		if e != nil {
			var err error
			contents, err = e.Encode()
			if err != nil {
				return fmt.Errorf("could not encode contents: %w", err)
			}
		}

		envelope := Envelope{
			BoardAddress: c.BoardAddress,
			Function:     f,
			Contents:     contents,
		}

		messageBytes, err := Encode(&envelope)
		if err != nil {
			return fmt.Errorf("could not encode envelope: %v", err)
		}
		bytesWritten, err := c.conn.Write(messageBytes)
		if err != nil {
			return fmt.Errorf("could not write message: %v", err)
		}
		logrus.Infof("Bytes written: %d", bytesWritten)
		if bytesWritten != len(messageBytes) {
			return fmt.Errorf("could not write full message; wrote %d bytes (expected: %d)", bytesWritten, len(messageBytes))
		}
	}

	{
		c.conn.SetDeadline(time.Now().Add(5 * time.Second))

		contents := make([]byte, 1024)
		bytesRead, err := c.conn.Read(contents)
		if err != nil {
			return fmt.Errorf("could not read contents: %v", err)
		}
		contents = contents[0:bytesRead]
		logrus.Infof("Bytes read: (%d) %x", bytesRead, contents)

		var envelope Envelope
		err = Decode(contents, &envelope)
		if err != nil {
			return fmt.Errorf("could not read contents: %v", err)
		}
		logrus.Debugf("Response: %x", envelope.Contents)

		if d != nil {
			err = Decode(envelope.Contents, d)
			if err != nil {
				return fmt.Errorf("could not decode response: %v", err)
			}
		}
	}

	return nil
}