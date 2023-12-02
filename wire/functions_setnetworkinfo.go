package wire

import (
	"net"
)

type SetNetworkInfoRequest struct {
	MACAddress net.HardwareAddr `wire:"length:6"`
	IPAddress  net.IP           `wire:"length:4"`
	Netmask    net.IP           `wire:"length:4"`
	Gateway    net.IP           `wire:"length:4"`
	Port       uint16
	_          [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}

type SetNetworkInfoResponse struct {
	Unknown1 uint8
	_        [0]byte `wire:"length:*"` // Fail if there are any leftover bytes.
}
