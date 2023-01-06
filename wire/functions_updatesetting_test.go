package wire

import (
	"testing"
)

func TestUpdateSetting(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "B0000A0000000000000000000000000000000000000000000000",
			output: UpdateSettingRequest{
				Address:  0xb0,
				Unknown1: 0x00,
				Value:    0x0a,
			},
		},
		{
			input:  "B0000A0000000000000000000000000000000000000000000001",
			output: UpdateSettingRequest{},
			fail:   true,
		},
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: UpdateSettingResponse{
				Result: 1,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
