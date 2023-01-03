package wire

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

type GetNetworkInfoRequest struct {
	Unknown1 uint8
}

func (r GetNetworkInfoRequest) Encode(writer *Writer) error {
	writer.WriteUint8(r.Unknown1)
	return nil
}

func (r *GetNetworkInfoRequest) Decode(reader *Reader) error {
	var err error
	r.Unknown1, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read unknown1: %v", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}

type GetNetworkInfoResponse struct {
	MACAddress net.HardwareAddr
	IPAddress  net.IP
	Netmask    net.IP
	Gateway    net.IP
	Port       uint16
}

func (r GetNetworkInfoResponse) Encode(writer *Writer) error {
	if len(r.MACAddress) != 6 {
		return fmt.Errorf("invalid MAC address size: %d (expected: 6)", len(r.MACAddress))
	}
	writer.WriteBytes(r.MACAddress)
	if len(r.IPAddress) != 4 {
		logrus.Debugf("IP address: %+v (%d)", r.IPAddress, len(r.IPAddress))
		return fmt.Errorf("invalid IP address size: %d (expected: 4)", len(r.IPAddress))
	}
	writer.WriteBytes(r.IPAddress)
	if len(r.Netmask) != 4 {
		return fmt.Errorf("invalid netmask size: %d (expected: 4)", len(r.Netmask))
	}
	writer.WriteBytes(r.Netmask)
	if len(r.Gateway) != 4 {
		return fmt.Errorf("invalid gateway address size: %d (expected: 4)", len(r.Gateway))
	}
	writer.WriteBytes(r.Gateway)
	writer.WriteUint16(r.Port)
	return nil
}

func (r *GetNetworkInfoResponse) Decode(reader *Reader) error {
	var err error
	r.MACAddress, err = reader.ReadBytes(6)
	if err != nil {
		return fmt.Errorf("could not read MAC address: %v", err)
	}
	r.IPAddress, err = reader.ReadBytes(4)
	if err != nil {
		return fmt.Errorf("could not read IP address: %v", err)
	}
	r.Netmask, err = reader.ReadBytes(4)
	if err != nil {
		return fmt.Errorf("could not read netmask: %v", err)
	}
	r.Gateway, err = reader.ReadBytes(4)
	if err != nil {
		return fmt.Errorf("could not read gateway: %v", err)
	}
	r.Port, err = reader.ReadUint16()
	if err != nil {
		return fmt.Errorf("could not read port: %v", err)
	}

	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", reader.Bytes())
	}
	return nil
}
