package wire

import (
	"testing"
)

func TestDeleteRecord(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "0100000070010308000000000000000000000000000000000000",
			output: DeleteRecordRequest{
				RecordIndex: 1,
				Unknown1:    []byte{0x70, 0x01, 0x03, 0x08},
			},
		},
		{
			input:  "0100000070010308000000000000000000000000000000000001",
			output: DeleteRecordRequest{},
			fail:   true,
		},
		{
			input: "0000000000000000000000000000000000000000000000000000",
			output: DeleteRecordResponse{
				Result: 0,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
