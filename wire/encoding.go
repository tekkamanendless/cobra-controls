package wire

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	TypeUint24   = "uint24"
	TypeDate     = "date"
	TypeTime     = "time"
	TypeDateTime = "datetime"
)

// Encoder can encode an object.
type Encoder interface {
	Encode() ([]byte, error)
}

// Decoder can decode an object.
type Decoder interface {
	Decode([]byte) error
}

// Encode an object to the wire.
func Encode(v any) ([]byte, error) {
	if e, ok := v.(Encoder); ok { // TODO: MOVE THIS TO encodeViaReflection
		return e.Encode()
	}
	return encodeViaReflection(v, encodingOptions{})
}

// Decode an object from the wire.
func Decode(data []byte, v any) error {
	if d, ok := v.(Decoder); ok { // TODO: MOVE THIS TO decodeViaReflection
		return d.Decode(data)
	}

	logrus.Debugf("Decode: v: %+v", v)
	return decodeViaReflection(NewReader(data), reflect.ValueOf(v), encodingOptions{})
}

// encodingOptions represents the options encoded in the "wire" field tag.
// type:{date,datetime,date,uint24}
// length:{*,[0-9]+}
type encodingOptions struct {
	Type   string
	Length int // -1 for "read everything until the end"
}

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
		case "length":
			if value == "*" {
				options.Length = -1
			} else {
				v, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return options, fmt.Errorf("invalid length: %q: %w", value, err)
				}
				options.Length = int(v)
			}
		case "type":
			switch value {
			case TypeDate, TypeDateTime, TypeTime, TypeUint24:
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

func encodeViaReflection(input any, options encodingOptions) ([]byte, error) {
	writer := NewWriter()

	myType := reflect.TypeOf(input)
	logrus.Debugf("encodeViaReflection: Kind: %v", myType.Kind())
	switch myType.Kind() {
	case reflect.Pointer:
		return encodeViaReflection(reflect.ValueOf(input).Elem().Interface(), options)
	case reflect.Struct:
		if timeValue, ok := input.(time.Time); ok {
			logrus.Debugf("encodeViaReflection: This is a time.Time.")
			if options.Type == "" {
				options.Type = TypeDateTime
			}
			switch options.Type {
			case TypeDate:
				writer.WriteDate(timeValue)
			case TypeTime:
				writer.WriteTime(timeValue)
			case TypeDateTime:
				writer.WriteDate(timeValue)
				writer.WriteTime(timeValue)
			default:
				return nil, fmt.Errorf("unhandled type: %s", options.Type)
			}
		} else {
			myValue := reflect.ValueOf(input)
			for f := 0; f < myType.NumField(); f++ {
				logrus.Debugf("encodeViaReflection: f: %d", f)
				myField := myType.Field(f)
				logrus.Debugf("encodeViaReflection: myField: %+v", myField)
				tag := myField.Tag.Get("wire")
				if tag == "-" {
					continue
				}
				options, err := parseOptionsFromTag(tag)
				if err != nil {
					return nil, err
				}
				var myFieldValue reflect.Value
				if myField.IsExported() {
					myFieldValue = myValue.Field(f)
				} else {
					myFieldValue = reflect.New(myField.Type)
				}
				contents, err := encodeViaReflection(myFieldValue.Interface(), options)
				if err != nil {
					return nil, err
				}
				writer.WriteBytes(contents)
			}
		}
	case reflect.Uint8:
		writer.WriteUint8(input.(uint8))
	case reflect.Uint16:
		writer.WriteUint16(input.(uint16))
	case reflect.Uint32:
		switch options.Type {
		case TypeUint24:
			writer.WriteUint24(input.(uint32))
		default:
			writer.WriteUint32(input.(uint32))
		}
	case reflect.Array:
		if myType.Elem().Kind() == reflect.Uint8 {
			logrus.Debugf("encodeViaReflection: array value: %+v", reflect.ValueOf(input))
			if myType.Len() == 0 {
				// Cool; this is always empty.
			} else {
				logrus.Debugf("encodeViaReflection: writing array of bytes: %x", reflect.ValueOf(input).Bytes())
				writer.WriteBytes(reflect.ValueOf(input).Bytes())
			}
		} else {
			return nil, fmt.Errorf("unimplemented: encoding an array of kind %q", myType.Elem().Kind())
		}
	case reflect.Slice:
		if myType.Elem().Kind() == reflect.Uint8 {
			logrus.Debugf("encodeViaReflection: slice value: %+v", reflect.ValueOf(input))

			logrus.Debugf("encodeViaReflection: writing slice of bytes: %x", reflect.ValueOf(input).Bytes())
			writer.WriteBytes(reflect.ValueOf(input).Bytes())
		} else {
			return nil, fmt.Errorf("unimplemented: encoding a slice of kind %q", myType.Elem().Kind())
		}
	default:
		return nil, fmt.Errorf("unimplemented: kind: %v", myType.Kind())
	}
	return writer.Bytes(), nil
}

func decodeViaReflection(reader *Reader, myValue reflect.Value, options encodingOptions) error {
	myType := myValue.Type()
	logrus.Debugf("decodeViaReflection: myValue: %+v", myValue)
	logrus.Debugf("decodeViaReflection: myType.Kind: %+v", myType.Kind())
	logrus.Debugf("decodeViaReflection: options: %+v", options)

	switch myType.Kind() {
	case reflect.Pointer:
		return decodeViaReflection(reader, myValue.Elem(), options)
	case reflect.Struct:
		if _, ok := myValue.Interface().(time.Time); ok {
			logrus.Debugf("decodeViaReflection: This is a time.Time.")
			if options.Type == "" {
				options.Type = TypeDateTime
			}
			switch options.Type {
			case TypeDate:
				v, err := reader.ReadDate()
				if err != nil {
					return fmt.Errorf("could not read date: %w", err)
				}
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf("could not set date: %w", err)
				}
			case TypeTime:
				v, err := reader.ReadTime()
				if err != nil {
					return fmt.Errorf("could not read time: %w", err)
				}
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf("could not set time: %w", err)
				}
			case TypeDateTime:
				v1, err := reader.ReadDate()
				if err != nil {
					return fmt.Errorf("could not read date: %w", err)
				}
				v2, err := reader.ReadTime()
				if err != nil {
					return fmt.Errorf("could not read time: %w", err)
				}
				v := MergeDateTime(v1, v2)
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf("could not set datetime: %w", err)
				}
			default:
				return fmt.Errorf("unhandled type: %s", options.Type)
			}
		} else {
			for f := 0; f < myType.NumField(); f++ {
				logrus.Debugf("decodeViaReflection: f: %d", f)
				myField := myType.Field(f)
				logrus.Debugf("decodeViaReflection: myField: %+v", myField)
				tag := myField.Tag.Get("wire")
				if tag == "-" {
					continue
				}
				options, err := parseOptionsFromTag(tag)
				if err != nil {
					return err
				}
				myFieldValue := myValue.Field(f)
				err = decodeViaReflection(reader, myFieldValue, options)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Array:
		logrus.Debugf("decodeViaReflection: array")
		if myType.Elem().Kind() == reflect.Uint8 {
			logrus.Debugf("decodeViaReflection: array of type uint8")

			if options.Length < 0 {
				options.Length = reader.Length()
			} else {
				options.Length = myType.Len()
			}

			contents, err := reader.ReadBytes(options.Length)
			if err != nil {
				return fmt.Errorf("could not read remainder: %w", err)
			}
			if myType.Len() == 0 {
				if IsAll(contents, 0) {
					logrus.Debugf("decodeViaReflection: contents is all 0.")
					contents = contents[len(contents):]
				} else {
					logrus.Debugf("decodeViaReflection: contents is not 0: %x", contents)
				}
			}
			if myType.Len() != len(contents) {
				return fmt.Errorf("unexpected length: %d (expected: %d)", len(contents), myType.Len())
			}
			for i := 0; i < myType.Len(); i++ {
				myValue.Index(i).Set(reflect.ValueOf(contents[i]))
			}
		} else {
			return fmt.Errorf("unhandled array type: %v", myType.Elem().Kind())
		}
	case reflect.Slice:
		logrus.Debugf("decodeViaReflection: slice")
		if myType.Elem().Kind() == reflect.Uint8 {
			logrus.Debugf("decodeViaReflection: slice of type uint8")

			contents, err := reader.ReadBytes(reader.Length())
			if err != nil {
				return fmt.Errorf("could not read remainder: %w", err)
			}
			var remainder []byte

			// A slice doesn't have the option of being "zero" length.  An un-length-limited
			// slice will grow to the entire length of the contents.
			if options.Length <= 0 {
				options.Length = len(contents)
			}
			if options.Length < len(contents) {
				remainder = contents[options.Length:]
				contents = contents[:options.Length]
				if IsAll(remainder, 0) {
					logrus.Debugf("decodeViaReflection: contents is all 0.")
				} else {
					logrus.Debugf("decodeViaReflection: contents is not 0: %x", contents)
				}
			}
			if options.Length != len(contents) {
				return fmt.Errorf("unexpected length: %d (expected: %d)", len(contents), options.Length)
			}
			logrus.Debugf("decodeViaReflection: making slice of length %d", options.Length)
			newValue := reflect.MakeSlice(myType, options.Length, options.Length)
			for i := 0; i < options.Length; i++ {
				logrus.Debugf("decodeViaReflection: setting index %d to 0x%x", i, contents[i])
				newValue.Index(i).Set(reflect.ValueOf(contents[i]))
			}
			myValue.Set(newValue)
		} else {
			return fmt.Errorf("unhandled slice type: %v", myType.Elem().Kind())
		}
	case reflect.Uint8:
		if options.Type != "" {
			return fmt.Errorf("invalid type for uint8: %q", options.Type)
		}
		v, err := reader.ReadUint8()
		if err != nil {
			return err
		}
		logrus.Debugf("decodeViaReflection: read uint8: %v", v)
		if myValue.CanUint() {
			myValue.SetUint(uint64(v))
		} else {
			return fmt.Errorf("could not set uint: %w", err)
		}
	case reflect.Uint16:
		if options.Type != "" {
			return fmt.Errorf("invalid type for uint16: %q", options.Type)
		}
		v, err := reader.ReadUint16()
		if err != nil {
			return err
		}
		logrus.Debugf("decodeViaReflection: read uint16: %v", v)
		if myValue.CanUint() {
			myValue.SetUint(uint64(v))
		} else {
			return fmt.Errorf("could not set uint: %w", err)
		}
	case reflect.Uint32:
		switch options.Type {
		case TypeUint24:
			v, err := reader.ReadUint24()
			if err != nil {
				return err
			}
			logrus.Debugf("decodeViaReflection: read uint24: %v", v)
			if myValue.CanUint() {
				myValue.SetUint(uint64(v))
			} else {
				return fmt.Errorf("could not set uint: %w", err)
			}
		case "":
			v, err := reader.ReadUint32()
			if err != nil {
				return err
			}
			logrus.Debugf("decodeViaReflection: read uint32: %v", v)
			if myValue.CanUint() {
				myValue.SetUint(uint64(v))
			} else {
				return fmt.Errorf("could not set uint: %w", err)
			}
		default:
			return fmt.Errorf("invalid type for uint8: %q", options.Type)
		}
	default:
		return fmt.Errorf("unhandled kind: %v", myType.Kind())
	}
	return nil
}
