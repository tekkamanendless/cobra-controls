package wire

import (
	"fmt"
	"net"
)

type GetNetworkInfoRequest struct {
	Unknown1 uint8
}

func (r *GetNetworkInfoRequest) Encode() ([]byte, error) {
	writer := NewWriter()
	writer.WriteUint8(r.Unknown1)
	return writer.Bytes(), nil
}

func (r *GetNetworkInfoRequest) Decode(b []byte) error {
	reader := NewReader(b)
	var err error
	r.Unknown1, err = reader.ReadUint8()
	if err != nil {
		return fmt.Errorf("could not read unknown1: %v", err)
	}
	if !IsAll(reader.Bytes(), 0) {
		return fmt.Errorf("unexpected contents: %x", b)
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

func (r *GetNetworkInfoResponse) Encode() ([]byte, error) {
	writer := NewWriter()
	if len(r.MACAddress) != 6 {
		return nil, fmt.Errorf("invalid MAC address size: %d (expected: 6)", len(r.MACAddress))
	}
	writer.WriteBytes(r.MACAddress)
	if len(r.IPAddress) != 4 {
		return nil, fmt.Errorf("invalid IP address size: %d (expected: 4)", len(r.IPAddress))
	}
	writer.WriteBytes(r.IPAddress)
	if len(r.Netmask) != 4 {
		return nil, fmt.Errorf("invalid netmaks size: %d (expected: 4)", len(r.Netmask))
	}
	writer.WriteBytes(r.Netmask)
	if len(r.Gateway) != 4 {
		return nil, fmt.Errorf("invalid gateway address size: %d (expected: 4)", len(r.Gateway))
	}
	writer.WriteBytes(r.Gateway)
	writer.WriteUint16(r.Port)
	return writer.Bytes(), nil
}

func (r *GetNetworkInfoResponse) Decode(b []byte) error {
	reader := NewReader(b)
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
		return fmt.Errorf("unexpected contents: %x", b)
	}
	return nil
}
