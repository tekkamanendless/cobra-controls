package wire

import (
	"testing"
	"time"
)

func TestGetOperationStatus(t *testing.T) {
	rows := []EncodeDecodeTest{
		{
			input:  "0000000000000000000000000000000000000000000000000000",
			output: GetOperationStatusRequest{},
		},
		{
			input:  "0100000000000000000000000000000000000000000000000001",
			output: GetOperationStatusRequest{},
			fail:   true,
		},
		{
			input: "221228031141419E290052018F5BB2009C2D955B00FF00000000",
			output: GetOperationStatusResponse{
				CurrentTime:   time.Date(2022, 12, 28, 11, 41, 41, 0, time.UTC),
				RecordCount:   10654,
				PopedomAmount: 338,
				Record: &Record{
					IDNumber:      23439,
					AreaNumber:    178,
					RecordState:   0,
					BrushDateTime: time.Date(2022, 12, 28, 11, 28, 42, 0, time.UTC),
				},
				RelayStatus: 0,
				MagnetState: 255,
				Reserved1:   0,
				FaultNumber: 0,
				Reserved2:   0,
				Reserved3:   0,
			},
		},
		{
			input: "221228031141419E29005201FFFFFFFFFFFFFFFF00FF00000000",
			output: GetOperationStatusResponse{
				CurrentTime:   time.Date(2022, 12, 28, 11, 41, 41, 0, time.UTC),
				RecordCount:   10654,
				PopedomAmount: 338,
				Record:        nil,
				RelayStatus:   0,
				MagnetState:   255,
				Reserved1:     0,
				FaultNumber:   0,
				Reserved2:     0,
				Reserved3:     0,
			},
		},
	}
	runEncodeDecodeTests(t, rows)
}
