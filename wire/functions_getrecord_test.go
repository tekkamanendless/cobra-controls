package wire

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestGetRecord(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	rows := []struct {
		input  string
		output any
		fail   bool
	}{
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: GetRecordRequest{
				RecordIndex: 1,
				//Remainder:   [0]uint8{},
			},
		},
		{
			input:  "0100000000000000000000000000000000000000000000000001",
			output: GetRecordRequest{},
			fail:   true,
		},
		{
			input: "6B9FBC02972D119170010308FFFFFFFFFFFFFFFF780103080000",
			output: GetRecordResponse{
				CardNumber:        40811,
				AreaNumber:        188,
				BrushCardState:    2,
				BrushCardDateTime: time.Date(2022, 12, 23, 18, 8, 34, 0, time.UTC),
				Unknown1:          []uint8{0x70, 0x01, 0x03, 0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x78, 0x01, 0x03, 0x08, 0x00, 0x00},
			},
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d", rowIndex), func(t *testing.T) {
			t.Run("Decode", func(t *testing.T) {
				var input []byte
				fmt.Sscanf(row.input, "%X", &input)
				require.Equal(t, row.input, fmt.Sprintf("%X", input))

				newValue := reflect.New(reflect.TypeOf(row.output))
				err := Decode(input, newValue.Interface())
				if row.fail {
					require.NotNil(t, err)
					return
				}
				require.Nil(t, err)
				assert.DeepEqual(t, row.output, newValue.Elem().Interface())
			})
			t.Run("Encode", func(t *testing.T) {
				if row.fail {
					return
				}

				var input []byte
				fmt.Sscanf(row.input, "%X", &input)
				require.Equal(t, row.input, fmt.Sprintf("%X", input))

				output, err := Encode(row.output)
				require.Nil(t, err)
				for len(output) < len(input) {
					output = append(output, 0x00)
				}
				assert.DeepEqual(t, input, output)
			})

		})
	}
}
