package wire

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Protocol string

const (
	ProtocolTCP = "tcp"
	ProtocolUDP = "udp"
)

const PortDefault uint16 = 60000

type Client struct {
	Protocol          Protocol
	ControllerAddress string
	ControllerPort    uint16
	BoardAddress      uint16

	BufferSize int

	conn net.Conn
}

func (c *Client) init() error {
	if len(c.Protocol) == 0 {
		c.Protocol = ProtocolTCP
	}
	if c.BufferSize == 0 {
		c.BufferSize = 1024
	}
	if c.ControllerPort == 0 {
		c.ControllerPort = PortDefault
	}

	if c.Protocol == ProtocolTCP && c.conn == nil {
		logrus.Debugf("Creating TCP connection.")

		var err error
		logrus.Debugf("Dialing (%s): %s:%d", c.Protocol, c.ControllerAddress, c.ControllerPort)
		c.conn, err = net.Dial(string(c.Protocol), fmt.Sprintf("%s:%d", c.ControllerAddress, c.ControllerPort))
		if err != nil {
			return err
		}
		logrus.Debugf("Connected via TCP.")
	}
	return nil
}

func (c *Client) RawUnicast(requestEnvelope Envelope) (*Envelope, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	if c.conn == nil {
		return nil, fmt.Errorf("no connection established")
	}

	{
		err := c.conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			return nil, err
		}

		messageWriter := NewWriter()
		err = Encode(messageWriter, &requestEnvelope)
		if err != nil {
			return nil, fmt.Errorf("could not encode envelope: %v", err)
		}
		bytesWritten, err := c.conn.Write(messageWriter.Bytes())
		if err != nil {
			c.conn.Close()
			c.conn = nil
			return nil, fmt.Errorf("could not write message: %v", err)
		}
		logrus.Debugf("Bytes written: %d", bytesWritten)
		if bytesWritten != messageWriter.Length() {
			return nil, fmt.Errorf("could not write full message; wrote %d bytes (expected: %d)", bytesWritten, messageWriter.Length())
		}
	}

	{
		err := c.conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			return nil, err
		}

		contents := make([]byte, c.BufferSize)
		bytesRead, err := c.conn.Read(contents)
		if err != nil {
			c.conn.Close()
			c.conn = nil
			return nil, fmt.Errorf("could not read contents: %v", err)
		}
		contents = contents[0:bytesRead]
		logrus.Debugf("Bytes read: (%d) %x", bytesRead, contents)

		reader := NewReader(contents)

		var responseEnvelope Envelope
		err = Decode(reader, &responseEnvelope)
		if err != nil {
			return nil, fmt.Errorf("could not decode envelope: %v", err)
		}
		logrus.Debugf("Response: %x", responseEnvelope.Contents)

		return &responseEnvelope, nil
	}
}

func (c *Client) RawMulticast(requestEnvelope Envelope) ([]*Envelope, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	var packetConns []net.PacketConn
	{
		ifaces, err := net.Interfaces()
		if err != nil {
			return nil, err
		}
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				return nil, err
			}
			logrus.Debugf("Interface %s has %d addresses: %s", iface.Name, len(addrs), addrs)
			if len(addrs) == 0 {
				continue
			}
			addr := addrs[0]

			ipAddress, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, err
			}
			if ipAddress.To4() == nil {
				logrus.Debugf("Skipping non-v4 IP address: %s", ipAddress)
				continue
			}
			logrus.Debugf("Creating UDP connection on: %s", ipAddress)

			logrus.Debugf("Listening (%s) on %s.", c.Protocol, ipAddress)
			packetConn, err := net.ListenPacket("udp", fmt.Sprintf("%s:0", ipAddress))
			if err != nil {
				return nil, err
			}
			defer packetConn.Close()
			logrus.Debugf("Connected via UDP on %s.", packetConn.LocalAddr())

			packetConns = append(packetConns, packetConn)
		}
	}

	var validPacketConns []net.PacketConn
	for _, packetConn := range packetConns {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.ControllerAddress, c.ControllerPort))
		if err != nil {
			return nil, err
		}

		messageWriter := NewWriter()
		err = Encode(messageWriter, &requestEnvelope)
		if err != nil {
			return nil, fmt.Errorf("could not encode envelope: %v", err)
		}
		bytesWritten, err := packetConn.WriteTo(messageWriter.Bytes(), addr)
		if err != nil {
			logrus.Debugf("Could not write message: [%T] %v", err, err)
			continue
		}
		logrus.Debugf("Bytes written: %d", bytesWritten)
		if bytesWritten != messageWriter.Length() {
			return nil, fmt.Errorf("could not write full message; wrote %d bytes (expected: %d)", bytesWritten, messageWriter.Length())
		}

		validPacketConns = append(validPacketConns, packetConn)
	}

	var packets [][]byte
	var mutex sync.Mutex
	var wg sync.WaitGroup
	for _, packetConn := range validPacketConns {
		wg.Add(1)
		go func(packetConn net.PacketConn) {
			defer wg.Done()

			logrus.Debugf("Reading packets from %s.", packetConn.LocalAddr())
			packetConn.SetDeadline(time.Now().Add(5 * time.Second))

			for {
				contents := make([]byte, c.BufferSize)
				bytesRead, sourceAddress, err := packetConn.ReadFrom(contents)
				if err != nil {
					if os.IsTimeout(err) {
						break
					}
					logrus.Errorf("Could not read contents: %v", err)
					return
				}
				_ = sourceAddress
				contents = contents[0:bytesRead]
				logrus.Debugf("Bytes read: (%d) %x", bytesRead, contents)

				mutex.Lock()
				packets = append(packets, contents)
				mutex.Unlock()
			}
		}(packetConn)
	}
	wg.Wait()
	logrus.Debugf("Read %d packets.", len(packets))

	var responseEnvelopes []*Envelope
	for i, contents := range packets {
		reader := NewReader(contents)

		var responseEnvelope Envelope
		err := Decode(reader, &responseEnvelope)
		if err != nil {
			return nil, fmt.Errorf("could not decode envelope %d: %v", i, err)
		}
		logrus.Debugf("Response %d: %x", i, responseEnvelope.Contents)

		responseEnvelopes = append(responseEnvelopes, &responseEnvelope)
	}

	return responseEnvelopes, nil
}

// Do performs a request and decodes the response.
func (c *Client) Do(functionCode uint16, request any, response any) error {
	_, err := c.DoWithEnvelopes(functionCode, request, response)
	return err
}

// DoWithEnvelopes performs a request and decodes the response.
// This will return the envelopes (with their full contents) as well.
func (c *Client) DoWithEnvelopes(functionCode uint16, request any, response any) ([]*Envelope, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	payloadWriter := NewWriter()
	if request != nil {
		err := Encode(payloadWriter, request)
		if err != nil {
			return nil, fmt.Errorf("could not encode contents: %w", err)
		}
	}

	requestEnvelope := Envelope{
		BoardAddress: c.BoardAddress,
		Function:     functionCode,
		Contents:     payloadWriter.Bytes(),
	}

	if c.Protocol == ProtocolTCP {
		responseEnvelope, err := c.RawUnicast(requestEnvelope)
		if err != nil {
			return nil, err
		}

		if response != nil {
			err = Decode(NewReader(responseEnvelope.Contents), response)
			if err != nil {
				return nil, fmt.Errorf("could not decode response: %w", err)
			}
		}

		return []*Envelope{responseEnvelope}, nil
	} else if c.Protocol == ProtocolUDP {
		responseEnvelopes, err := c.RawMulticast(requestEnvelope)
		if err != nil {
			return nil, err
		}

		myValue := reflect.ValueOf(response)
		if myValue.Type().Kind() == reflect.Pointer {
			logrus.Debugf("Initial response type: %+v", myValue.Type())
			if myValue.IsNil() {
				logrus.Debugf("Initial value is nil.")
				newValue := reflect.New(myValue.Type().Elem())
				myValue = newValue
			}
			myValue = myValue.Elem()
		}
		logrus.Debugf("Response type: %v", myValue.Type())

		readMany := false
		if myValue.Type().Kind() == reflect.Array {
			logrus.Debugf("This is an array.")
			readMany = true
		} else if myValue.Type().Kind() == reflect.Slice {
			logrus.Debugf("This is a slice.")
			readMany = true
		}
		logrus.Debugf("Read many?: %t", readMany)

		if !readMany {
			if len(responseEnvelopes) == 0 {
				return nil, os.ErrDeadlineExceeded
			}
			responseEnvelope := responseEnvelopes[0]

			if response != nil {
				err = Decode(NewReader(responseEnvelope.Contents), response)
				if err != nil {
					return nil, fmt.Errorf("could not decode response: %v", err)
				}
			}
		} else {
			if myValue.Type().Kind() == reflect.Slice {
				myValue.Set(reflect.MakeSlice(myValue.Type(), len(responseEnvelopes), len(responseEnvelopes)))
			}

			for i, responseEnvelope := range responseEnvelopes {
				if response != nil {
					if i >= myValue.Cap() {
						logrus.Debugf("Hit the capacity of the array or slice: %d", myValue.Cap())
						break
					}
					err = Decode(NewReader(responseEnvelope.Contents), myValue.Index(i).Addr().Interface())
					if err != nil {
						return nil, fmt.Errorf("could not decode response %d: %v", i, err)
					}
				}
			}
		}

		return responseEnvelopes, nil
	}

	return nil, fmt.Errorf("invalid protocol: %s", c.Protocol)
}
