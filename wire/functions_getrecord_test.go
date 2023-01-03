package wire

import (
	"testing"
	"time"
)

func TestGetRecord(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: GetRecordRequest{
				RecordIndex: 1,
				//Remainder:   [0]uint8{},
			},
		},
		{
			input:  "0100000000000000000000000000000000000000000000000001",
			output: GetRecordRequest{},
			fail:   true,
		},
		{
			input: "6B9FBC02972D119170010308FFFFFFFFFFFFFFFF780103080000",
			output: GetRecordResponse{
				CardNumber:        40811,
				AreaNumber:        188,
				BrushCardState:    2,
				BrushCardDateTime: time.Date(2022, 12, 23, 18, 8, 34, 0, time.UTC),
				Unknown1:          []uint8{0x70, 0x01, 0x03, 0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x78, 0x01, 0x03, 0x08, 0x00, 0x00},
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
