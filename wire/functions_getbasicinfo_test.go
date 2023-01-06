package wire

import (
	"testing"
	"time"
)

func TestGetBasicInfo(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input:  "0000000000000000000000000000000000000000000000000000",
			output: GetBasicInfoRequest{},
		},
		{
			input:  "0100000000000000000000000000000000000000000000000001",
			output: GetBasicInfoRequest{},
			fail:   true,
		},
		{
			input: "0810061E64012401CFFFF0FFFFFFFF0000000000000064887400",
			output: GetBasicInfoResponse{
				IssueDate: time.Date(2008, 10, 6, 0, 0, 0, 0, time.UTC),
				Version:   30,
				Model:     100,
				Unknown1:  []uint8{0x01, 0x24, 0x01, 0xcf, 0xff, 0xf0, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64, 0x88, 0x74, 0x00},
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
