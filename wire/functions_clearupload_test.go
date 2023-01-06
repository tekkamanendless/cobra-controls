package wire

import (
	"testing"
)

func TestClearUpload(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input:  "0000000000000000000000000000000000000000000000000000",
			output: ClearUploadRequest{},
		},
		{
			input:  "0000000000000000000000000000000000000000000000000001",
			output: ClearUploadRequest{},
			fail:   true,
		},
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: ClearUploadResponse{
				Result: 1,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
