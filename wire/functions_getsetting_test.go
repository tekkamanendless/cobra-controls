package wire

import (
	"testing"
)

func TestGetSetting(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "1C00000000000000000000000000000000000000000000000000",
			output: GetSettingRequest{
				Address:  28,
				Unknown1: 0,
			},
		},
		{
			input:  "1C00000000000000000000000000000000000000000000000001",
			output: GetSettingRequest{},
			fail:   true,
		},
		{
			input: "0303000000000102030400000000FF0000000000000000000000",
			output: GetSettingResponse{
				Value:    3,
				Unknown1: []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
