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

const PortDefault = 60000

type Client struct {
	Protocol          Protocol
	ControllerAddress string
	ControllerPort    int
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

func (c *Client) Raw(f uint16, request any, response any) error {
	if err := c.init(); err != nil {
		return err
	}

	if c.Protocol == ProtocolTCP {
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
	} else if c.Protocol == ProtocolUDP {
		var packetConns []net.PacketConn
		{
			ifaces, err := net.Interfaces()
			if err != nil {
				return err
			}
			for _, iface := range ifaces {
				addrs, err := iface.Addrs()
				if err != nil {
					return err
				}
				logrus.Debugf("Interface %s has %d addresses: %s", iface.Name, len(addrs), addrs)
				if len(addrs) == 0 {
					continue
				}
				addr := addrs[0]

				ipAddress, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					return err
				}
				if ipAddress.To4() == nil {
					logrus.Debugf("Skipping non-v4 iP address: %s", ipAddress)
					continue
				}
				logrus.Debugf("Creating UDP connection on: %s", ipAddress)

				logrus.Debugf("Listening (%s) on %s.", c.Protocol, ipAddress)
				packetConn, err := net.ListenPacket("udp", fmt.Sprintf("%s:0", ipAddress))
				if err != nil {
					return err
				}
				defer packetConn.Close()
				logrus.Debugf("Connected via UDP on %s.", packetConn.LocalAddr())

				packetConns = append(packetConns, packetConn)
			}
		}

		for _, packetConn := range packetConns {
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

			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.ControllerAddress, c.ControllerPort))
			if err != nil {
				return err
			}

			bytesWritten, err := packetConn.WriteTo(messageWriter.Bytes(), addr)
			if err != nil {
				return fmt.Errorf("could not write message: %v", err)
			}
			logrus.Debugf("Bytes written: %d", bytesWritten)
			if bytesWritten != messageWriter.Length() {
				return fmt.Errorf("could not write full message; wrote %d bytes (expected: %d)", bytesWritten, messageWriter.Length())
			}
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

		var packets [][]byte
		var mutex sync.Mutex
		var wg sync.WaitGroup
		for _, packetConn := range packetConns {
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

					if !readMany {
						break
					}
				}
			}(packetConn)
		}
		wg.Wait()
		logrus.Debugf("Read %d packets.", len(packets))

		if !readMany {
			if len(packets) == 0 {
				return os.ErrDeadlineExceeded
			}

			contents := packets[0]

			reader := NewReader(contents)

			var envelope Envelope
			err := Decode(reader, &envelope)
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
		} else {
			if myValue.Type().Kind() == reflect.Slice {
				myValue.Set(reflect.MakeSlice(myValue.Type(), len(packets), len(packets)))
			}

			for i, contents := range packets {
				reader := NewReader(contents)

				var envelope Envelope
				err := Decode(reader, &envelope)
				if err != nil {
					return fmt.Errorf("could not decode envelope %d: %v", i, err)
				}
				logrus.Debugf("Response %d: %x", i, envelope.Contents)

				if response != nil {
					if i >= myValue.Cap() {
						logrus.Debugf("Hit the capacity of the array or slice: %d", myValue.Cap())
						break
					}
					err = Decode(NewReader(envelope.Contents), myValue.Index(i).Addr().Interface())
					if err != nil {
						return fmt.Errorf("could not decode response %d: %v", i, err)
					}
				}
			}
		}
	}

	return nil
}
