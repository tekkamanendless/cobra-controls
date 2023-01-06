package wire

import (
	"testing"
)

func TestUnknown1098(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input:  "0000000000000000000000000000000000000000000000000000",
			output: Unknown1098Request{},
		},
		{
			input:  "0000000000000000000000000000000000000000000000000001",
			output: Unknown1098Request{},
			fail:   true,
		},
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: Unknown1098Response{
				Result: 1,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
