package wire

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestEnvelope(t *testing.T) {
	{
		var data []byte
		fmt.Sscanf("7E57F282100000000000000000000000000000000000000000000000000000DB010D", "%X", &data)

		var envelope Envelope
		err := Decode(NewReader(data), &envelope)
		require.Nil(t, err)

		assert.Equal(t, uint16(0xF257), envelope.BoardAddress)
		assert.Equal(t, uint16(0x1082), envelope.Function)
		assert.Equal(t, 26, len(envelope.Contents))

		writer := NewWriter()
		err = Encode(writer, &envelope)
		require.Nil(t, err)
		require.Equal(t, data, writer.Bytes())
	}
}
