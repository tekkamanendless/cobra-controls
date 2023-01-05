package wire

import (
	"fmt"
	"strconv"
	"strings"
)

// encodingOptions represents the options encoded in the "wire" field tag.
// type:{date,datetime,date,uint24}
// length:{*,[0-9]+}
type encodingOptions struct {
	Name   string // This is primarily passed in to help with rendering a meaningful error message.
	Type   string // This is a Type* constant for the type of the field.
	Length int    // -1 for "read everything until the end", otherwise this is the actual length.  "0" is acceptable.
	Null   *uint8 // If this value is null, fill in this byte for the length (typically 0xff).
}

const (
	OptionLength = "length"
	OptionNull   = "null"
	OptionType   = "type"
)

// parseOptionsFromTag parses the "wire" tag and returns the "encodingOptions" for it.
func parseOptionsFromTag(tag string) (encodingOptions, error) {
	options := encodingOptions{}
	if tag == "" {
		return options, nil
	}
	items := strings.Split(tag, ",")
	for _, item := range items {
		parts := strings.SplitN(item, ":", 2)
		if len(parts) != 2 {
			return options, fmt.Errorf("invalid option: %q", item)
		}
		key := parts[0]
		value := parts[1]
		switch key {
		case OptionLength:
			if value == "*" {
				options.Length = -1
			} else {
				v, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return options, fmt.Errorf("invalid length: %q: %w", value, err)
				}
				options.Length = int(v)
			}
		case OptionNull:
			v, err := strconv.ParseInt(value, 0 /*auto-detect*/, 9)
			if err != nil {
				return options, fmt.Errorf("invalid byte: %q: %w", value, err)
			}
			nullValue := new(uint8)
			*nullValue = uint8(v)
			options.Null = nullValue
		case OptionType:
			switch value {
			case TypeDate, TypeDateTime, TypeHexDate, TypeHexDateTime, TypeHexTime, TypeTime, TypeUint24:
				// This is valid.
			default:
				return options, fmt.Errorf("invalid type: %q", value)
			}
			options.Type = value
		default:
			return options, fmt.Errorf("invalid key: %q", key)
		}
	}
	return options, nil
}
