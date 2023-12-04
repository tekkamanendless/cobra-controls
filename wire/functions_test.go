package wire

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EncodeDecodeTest struct {
	input  string
	output any
	fail   bool
}

func runEncodeDecodeTests(t *testing.T, rows []EncodeDecodeTest) {
	logrus.SetLevel(logrus.DebugLevel)
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d", rowIndex), func(t *testing.T) {
			t.Run(fmt.Sprintf("%T", row.output), func(t *testing.T) {
				t.Run("Decode", func(t *testing.T) {
					var input []byte
					fmt.Sscanf(row.input, "%X", &input)
					require.Equal(t, row.input, fmt.Sprintf("%X", input))

					newValue := reflect.New(reflect.TypeOf(row.output))
					err := Decode(NewReader(input), newValue.Interface())
					if row.fail {
						require.NotNil(t, err)
						return
					}
					require.Nil(t, err)
					assert.Equal(t, row.output, newValue.Elem().Interface())
				})
				t.Run("Encode", func(t *testing.T) {
					if row.fail {
						return
					}

					var input []byte
					fmt.Sscanf(row.input, "%X", &input)
					require.Equal(t, row.input, fmt.Sprintf("%X", input))

					writer := NewWriter()
					err := Encode(writer, row.output)
					require.Nil(t, err)
					for len(writer.Bytes()) < len(input) {
						writer.WriteUint8(0)
					}
					assert.Equal(t, input, writer.Bytes())
				})
			})
		})
	}
}
