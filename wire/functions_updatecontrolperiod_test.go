package wire

import (
	"testing"
	"time"
)

func TestUpdateControlPeriod(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input: "020008020000009000B000000000000000008620612100000000",
			output: UpdateControlPeriodRequest{
				TimeIndex:         2,
				WeekControl:       8,
				NextLinkTimeIndex: 2,
				Standby1:          0,
				Standby2:          0,
				StartTime1:        time.Date(0, time.January, 1, 18, 0, 0, 0, time.UTC),
				EndTime1:          time.Date(0, time.January, 1, 22, 0, 0, 0, time.UTC),
				StartTime2:        time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime2:          time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				StartTime3:        time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime3:          time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				StartDate:         time.Date(2016, time.April, 6, 0, 0, 0, 0, time.UTC),
				EndDate:           time.Date(2016, time.November, 1, 0, 0, 0, 0, time.UTC),
				Standby3:          0,
				Standby4:          0,
				Standby5:          0,
				Standby6:          0,
			},
		},
		{
			input:  "020008020000009000B000000000000000008620612100000000" + "01",
			output: UpdateControlPeriodRequest{},
			fail:   true,
		},
		{
			input: "020008020000009000B000000000000000008620612100000000",
			output: UpdateControlPeriodResponse{
				TimeIndex:         2,
				WeekControl:       8,
				NextLinkTimeIndex: 2,
				Standby1:          0,
				Standby2:          0,
				StartTime1:        time.Date(0, time.January, 1, 18, 0, 0, 0, time.UTC),
				EndTime1:          time.Date(0, time.January, 1, 22, 0, 0, 0, time.UTC),
				StartTime2:        time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime2:          time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				StartTime3:        time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime3:          time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
				StartDate:         time.Date(2016, time.April, 6, 0, 0, 0, 0, time.UTC),
				EndDate:           time.Date(2016, time.November, 1, 0, 0, 0, 0, time.UTC),
				Standby3:          0,
				Standby4:          0,
				Standby5:          0,
				Standby6:          0,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
