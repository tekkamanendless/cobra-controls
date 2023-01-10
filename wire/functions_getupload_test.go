package wire

import (
	"testing"
	"time"
)

func TestGetUpload(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: GetUploadRequest{
				Index: 1,
			},
		},
		{
			input:  "0100000000000000000000000000000000000000000000000001",
			output: GetUploadRequest{},
			fail:   true,
		},
		{
			input: "C09D0B0121009F65010000000000000000000000000000000000",
			output: GetUploadResponse{
				IDNumber:   40384,
				AreaNumber: 11,
				DoorNumber: 1,
				StartDate:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2050, 12, 31, 0, 0, 0, 0, time.UTC),
				Time:       1,
				Password:   0,
				Standby1:   0,
				Standby2:   0,
				Standby3:   0,
				Standby4:   0,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
