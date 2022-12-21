package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestParseDate(t *testing.T) {
	data := []byte{0xB3, 0x02}
	output, err := parseDate(data)
	require.Nil(t, err)
	assert.Equal(t, time.Date(2001, time.May, 19, 0, 0, 0, 0, time.UTC), output)
}

func TestParseTime(t *testing.T) {
	data := []byte{0x0D, 0x5D}
	output, err := parseTime(data)
	require.Nil(t, err)
	assert.Equal(t, time.Date(0, time.January, 1, 11, 40, 26, 0, time.UTC), output)
}
