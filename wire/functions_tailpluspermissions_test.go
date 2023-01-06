package wire

import (
	"testing"
	"time"
)

func TestTailPlusPermissions(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "3A03618EC90421009F6501000000000000000000000000000000",
			output: TailPlusPermissionsRequest{
				UploadIndex: 826,
				CardNumber:  36449,
				AreaNumber:  201,
				Door:        4,
				StartDate:   time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2050, 12, 31, 0, 0, 0, 0, time.UTC),
				Time:        1,
				Password:    0,
				Standby1:    0,
				Standby2:    0,
				Standby3:    0,
				Standby4:    0,
			},
		},
		{
			input:  "3A03618EC90421009F6501000000000000000000000000000001",
			output: TailPlusPermissionsRequest{},
			fail:   true,
		},
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: TailPlusPermissionsResponse{
				Result: 1,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
