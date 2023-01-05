package wire

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEncoding(t *testing.T) {
	t.Run("Time", func(t *testing.T) {
		t.Run("16-bit", func(t *testing.T) {
			//0b0000110000100010 = 0x0c22 (date)
			//0b0111100010000011 = 0x7883 (time); note that we can't do 5s because of the /2 encoding.

			t.Run("Default is 16-bit datetime", func(t *testing.T) {
				type MyStruct struct {
					Time1 time.Time
				}

				t.Run("Decode", func(t *testing.T) {
					input := []byte{0x22, 0x0c, 0x83, 0x78}
					var output MyStruct
					err := Decode(NewReader(input), &output)
					require.Nil(t, err)
					require.Equal(t, time.Date(2006, 01, 02, 15, 4, 6, 0, time.UTC), output.Time1)
				})
				t.Run("Encode", func(t *testing.T) {
					input := MyStruct{
						Time1: time.Date(2006, 01, 02, 15, 4, 6, 0, time.UTC),
					}
					writer := NewWriter()
					err := Encode(writer, input)
					require.Nil(t, err)
					require.Equal(t, []byte{0x22, 0x0c, 0x83, 0x78}, writer.Bytes())
				})
			})
			t.Run("datetime", func(t *testing.T) {
				type MyStruct struct {
					Time1 time.Time `wire:"type:datetime"`
				}

				t.Run("Decode", func(t *testing.T) {
					input := []byte{0x22, 0x0c, 0x83, 0x78}
					var output MyStruct
					err := Decode(NewReader(input), &output)
					require.Nil(t, err)
					require.Equal(t, time.Date(2006, 01, 02, 15, 4, 6, 0, time.UTC), output.Time1)
				})
				t.Run("Encode", func(t *testing.T) {
					input := MyStruct{
						Time1: time.Date(2006, 01, 02, 15, 4, 6, 0, time.UTC),
					}
					writer := NewWriter()
					err := Encode(writer, input)
					require.Nil(t, err)
					require.Equal(t, []byte{0x22, 0x0c, 0x83, 0x78}, writer.Bytes())
				})
			})
			t.Run("date", func(t *testing.T) {
				type MyStruct struct {
					Time1 time.Time `wire:"type:date"`
				}

				t.Run("Decode", func(t *testing.T) {
					input := []byte{0x22, 0x0c}
					var output MyStruct
					err := Decode(NewReader(input), &output)
					require.Nil(t, err)
					require.Equal(t, time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC), output.Time1)
				})
				t.Run("Encode", func(t *testing.T) {
					input := MyStruct{
						Time1: time.Date(2006, 01, 02, 15, 4, 6, 0, time.UTC),
					}
					writer := NewWriter()
					err := Encode(writer, input)
					require.Nil(t, err)
					require.Equal(t, []byte{0x22, 0x0c}, writer.Bytes())
				})
			})
			t.Run("time", func(t *testing.T) {
				type MyStruct struct {
					Time1 time.Time `wire:"type:time"`
				}

				t.Run("Decode", func(t *testing.T) {
					input := []byte{0x83, 0x78}
					var output MyStruct
					err := Decode(NewReader(input), &output)
					require.Nil(t, err)
					require.Equal(t, time.Date(0, 01, 01, 15, 4, 6, 0, time.UTC), output.Time1)
				})
				t.Run("Encode", func(t *testing.T) {
					input := MyStruct{
						Time1: time.Date(2006, 01, 02, 15, 4, 6, 0, time.UTC),
					}
					writer := NewWriter()
					err := Encode(writer, input)
					require.Nil(t, err)
					require.Equal(t, []byte{0x83, 0x78}, writer.Bytes())
				})
			})
		})
		t.Run("Hex", func(t *testing.T) {
			t.Run("datetime", func(t *testing.T) {
				type MyStruct struct {
					Time1 time.Time `wire:"type:hexdatetime"`
				}

				t.Run("Decode", func(t *testing.T) {
					input := []byte{0x06, 0x01, 0x02, 0x01, 0x15, 0x04, 0x05}
					var output MyStruct
					err := Decode(NewReader(input), &output)
					require.Nil(t, err)
					require.Equal(t, time.Date(2006, 01, 02, 15, 4, 5, 0, time.UTC), output.Time1)
				})
				t.Run("Encode", func(t *testing.T) {
					input := MyStruct{
						Time1: time.Date(2006, 01, 02, 15, 4, 5, 0, time.UTC),
					}
					writer := NewWriter()
					err := Encode(writer, input)
					require.Nil(t, err)
					require.Equal(t, []byte{0x06, 0x01, 0x02, 0x01, 0x15, 0x04, 0x05}, writer.Bytes())
				})
			})
			t.Run("date", func(t *testing.T) {
				type MyStruct struct {
					Time1 time.Time `wire:"type:hexdate"`
				}

				t.Run("Decode", func(t *testing.T) {
					input := []byte{0x06, 0x01, 0x02}
					var output MyStruct
					err := Decode(NewReader(input), &output)
					require.Nil(t, err)
					require.Equal(t, time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC), output.Time1)
				})
				t.Run("Encode", func(t *testing.T) {
					input := MyStruct{
						Time1: time.Date(2006, 01, 02, 15, 4, 6, 0, time.UTC),
					}
					writer := NewWriter()
					err := Encode(writer, input)
					require.Nil(t, err)
					require.Equal(t, []byte{0x06, 0x01, 0x02}, writer.Bytes())
				})
			})
			t.Run("time", func(t *testing.T) {
				type MyStruct struct {
					Time1 time.Time `wire:"type:hextime"`
				}

				t.Run("Decode", func(t *testing.T) {
					input := []byte{0x15, 0x04, 0x05}
					var output MyStruct
					err := Decode(NewReader(input), &output)
					require.Nil(t, err)
					require.Equal(t, time.Date(0, 01, 01, 15, 4, 5, 0, time.UTC), output.Time1)
				})
				t.Run("Encode", func(t *testing.T) {
					input := MyStruct{
						Time1: time.Date(2006, 01, 02, 15, 4, 5, 0, time.UTC),
					}
					writer := NewWriter()
					err := Encode(writer, input)
					require.Nil(t, err)
					require.Equal(t, []byte{0x15, 0x04, 0x05}, writer.Bytes())
				})
			})
		})
	})
	t.Run("Options", func(t *testing.T) {
		t.Run("Length has to be an integer", func(t *testing.T) {
			t.Run("3.5", func(t *testing.T) {
				type MyStruct struct {
					Key1 []byte `wire:"length:3.5"`
				}
				input := MyStruct{}
				writer := NewWriter()
				err := Encode(writer, input)
				require.NotNil(t, err)
			})
			t.Run("empty", func(t *testing.T) {
				type MyStruct struct {
					Key1 []byte `wire:"length:"`
				}
				input := MyStruct{}
				writer := NewWriter()
				err := Encode(writer, input)
				require.NotNil(t, err)
			})
			t.Run("BOGUS", func(t *testing.T) {
				type MyStruct struct {
					Key1 []byte `wire:"length:BOGUS"`
				}
				input := MyStruct{}
				writer := NewWriter()
				err := Encode(writer, input)
				require.NotNil(t, err)
			})
		})
	})
}
