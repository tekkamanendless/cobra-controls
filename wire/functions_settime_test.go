package wire

import (
	"testing"
	"time"
)

func TestSetTime(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "2212230521385000000000000000000000000000000000000000",
			output: SetTimeRequest{
				CurrentTime: time.Date(2022, 12, 23, 21, 38, 50, 0, time.UTC),
			},
		},
		{
			input:  "2212230521385000000000000000000000000000000000000001",
			output: SetTimeRequest{},
			fail:   true,
		},
		{
			input: "2212230521385000000000000000000000000000000000000000",
			output: SetTimeResponse{
				CurrentTime: time.Date(2022, 12, 23, 21, 38, 50, 0, time.UTC),
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
