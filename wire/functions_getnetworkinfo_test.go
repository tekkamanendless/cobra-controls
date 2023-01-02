package wire

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestGetNetworkInfo(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	rows := []struct {
		input  string
		output any
		fail   bool
	}{
		{
			input: "0100000000000000000000000000000000000000000000000000",
			output: GetNetworkInfoRequest{
				Unknown1: 1,
			},
		},
		{
			input:  "0100000000000000000000000000000000000000000000000001",
			output: GetNetworkInfoRequest{},
			fail:   true,
		},
		{
			input: "00574764F010C0A8C9C2FFFFFF00C0A8C9FE60EA000000000000",
			output: GetNetworkInfoResponse{
				MACAddress: net.HardwareAddr([]byte{0x00, 0x57, 0x47, 0x64, 0xf0, 0x10}),
				IPAddress:  net.IP([]byte{192, 168, 201, 194}),
				Netmask:    net.IP([]byte{255, 255, 255, 0}),
				Gateway:    net.IP([]byte{192, 168, 201, 254}),
				Port:       60000,
			},
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d", rowIndex), func(t *testing.T) {
			t.Run(fmt.Sprintf("%T", row.output), func(t *testing.T) {
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
		})
	}
}
