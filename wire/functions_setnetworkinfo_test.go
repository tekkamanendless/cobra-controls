package wire

import (
	"net"
	"testing"
)

func TestSetNetworkInfo(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "00574764F010C0A8C9C2FFFFFF00C0A8C9FE60EA000000000000",
			output: SetNetworkInfoRequest{
				MACAddress: net.HardwareAddr([]byte{0x00, 0x57, 0x47, 0x64, 0xf0, 0x10}),
				IPAddress:  net.IP([]byte{192, 168, 201, 194}),
				Netmask:    net.IP([]byte{255, 255, 255, 0}),
				Gateway:    net.IP([]byte{192, 168, 201, 254}),
				Port:       60000,
			},
		},
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: SetNetworkInfoResponse{
				Unknown1: 1,
			},
		},
		{
			input:  "0100000000000000000000000000000000000000000000000001",
			output: SetNetworkInfoResponse{},
			fail:   true,
		},
	}
	runEncodeDecodeTests(t, rows)
}
