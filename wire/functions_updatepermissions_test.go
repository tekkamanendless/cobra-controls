package wire

import (
	"testing"
	"time"
)

func TestUpdatePermissions(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "01007028530121009F650140E201000000000000000000000000",
			output: UpdatePermissionsRequest{
				Unknown1:  1,
				CardID:    10352,
				Area:      83,
				Door:      1,
				StartDate: time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2050, 12, 31, 0, 0, 0, 0, time.UTC),
				Time:      1,
				Password:  123456,
				Standby:   []byte{0, 0, 0, 0},
			},
		},
		{
			input:  "01007028530121009F650140E201000000000000000000000001",
			output: UpdatePermissionsRequest{},
			fail:   true,
		},
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: UpdatePermissionsResponse{
				Result: 1,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
