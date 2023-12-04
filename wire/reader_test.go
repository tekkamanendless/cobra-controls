package wire

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDate(t *testing.T) {
	data := []byte{0xB3, 0x02}
	r := NewReader(data)
	output, err := r.ReadDate()
	require.Nil(t, err)
	assert.Equal(t, time.Date(2001, time.May, 19, 0, 0, 0, 0, time.UTC), output)
}

func TestReadTime(t *testing.T) {
	data := []byte{0x0D, 0x5D}
	r := NewReader(data)
	output, err := r.ReadTime()
	require.Nil(t, err)
	assert.Equal(t, time.Date(0, time.January, 1, 11, 40, 26, 0, time.UTC), output)
}

func TestReadUint8(t *testing.T) {
	data := []byte{0x01}
	r := NewReader(data)
	output, err := r.ReadUint8()
	require.Nil(t, err)
	assert.Equal(t, uint8(1), output)
}

func TestReadUint16(t *testing.T) {
	data := []byte{0x01, 0x02}
	r := NewReader(data)
	output, err := r.ReadUint16()
	require.Nil(t, err)
	assert.Equal(t, uint16(513), output)
}

func TestReadUint24(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	r := NewReader(data)
	output, err := r.ReadUint24()
	require.Nil(t, err)
	assert.Equal(t, uint32(197121), output)
}

func TestReadUint32(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	r := NewReader(data)
	output, err := r.ReadUint32()
	require.Nil(t, err)
	assert.Equal(t, uint32(67305985), output)
}
