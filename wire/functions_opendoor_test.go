package wire

import (
	"testing"
)

func TestOpenDoor(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "0401000000000000000000000000000000000000000000000000",
			output: OpenDoorRequest{
				Door:     4,
				Unkonwn1: 1,
			},
		},
		{
			input:  "0401000000000000000000000000000000000000000000000001",
			output: OpenDoorRequest{},
			fail:   true,
		},
		{
			input:  "0000000000000000000000000000000000000000000000000000",
			output: OpenDoorResponse{},
		},
	}
	runEncodeDecodeTests(t, rows)
}
