package wire

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestCardID(t *testing.T) {
	rows := []struct {
		prefix uint8
		suffix uint16
		output string
	}{
		{
			prefix: 0,
			suffix: 0,
			output: "000000",
		},
		{
			prefix: 1,
			suffix: 1,
			output: "100001",
		},
		{
			prefix: 2,
			suffix: 2,
			output: "200002",
		},
		{
			prefix: 250,
			suffix: 10,
			output: "25000010",
		},
		{
			prefix: 250,
			suffix: 65000,
			output: "25065000",
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d/%s", rowIndex, row.output), func(t *testing.T) {
			output := CardID(row.prefix, row.suffix)
			assert.Equal(t, row.output, output)
		})
	}

}

func TestInsaneBase16ToBase10(t *testing.T) {
	rows := []struct {
		input  uint8
		output uint8
	}{
		{
			input:  0x0,
			output: 0,
		},
		{
			input:  0x1,
			output: 1,
		},
		{
			input:  0x9,
			output: 9,
		},
		{
			input:  0x10,
			output: 10,
		},
		{
			input:  0x11,
			output: 11,
		},
		{
			input:  0x22,
			output: 22,
		},
		{
			input:  0x59,
			output: 59,
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d/0x%x", rowIndex, row.input), func(t *testing.T) {
			output := InsaneBase16ToBase10(row.input)
			assert.Equal(t, row.output, output)
		})
	}
}

func TestInsaneBase10ToBase16(t *testing.T) {
	rows := []struct {
		input  uint8
		output uint8
	}{
		{
			input:  0,
			output: 0x0,
		},
		{
			input:  1,
			output: 0x1,
		},
		{
			input:  9,
			output: 0x9,
		},
		{
			input:  10,
			output: 0x10,
		},
		{
			input:  11,
			output: 0x11,
		},
		{
			input:  22,
			output: 0x22,
		},
		{
			input:  59,
			output: 0x59,
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d/0x%x", rowIndex, row.input), func(t *testing.T) {
			output := InsaneBase10ToBase16(row.input)
			assert.Equal(t, row.output, output)
		})
	}

}
