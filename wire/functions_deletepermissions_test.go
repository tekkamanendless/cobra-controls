package wire

import (
	"testing"
	"time"
)

func TestDeletePermissions(t *testing.T) {
	timePointer := func(t time.Time) *time.Time {
		return &t
	}

	rows := []EncodeDecodeTest{
		{
			input: "00007028530121009F650140E201000000000000000000000000",
			output: DeletePermissionsRequest{
				Empty1:    0,
				CardID:    10352,
				Area:      83,
				Door:      1,
				StartDate: timePointer(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:   timePointer(time.Date(2050, 12, 31, 0, 0, 0, 0, time.UTC)),
				Time:      1,
				Password:  123456,
				Standby:   []byte{0, 0, 0, 0},
			},
		},
		{
			input:  "00007028530121009F650140E201000000000000000000000001",
			output: DeletePermissionsRequest{},
			fail:   true,
		},
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: DeletePermissionsResponse{
				Result: 1,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
