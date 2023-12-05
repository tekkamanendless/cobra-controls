package wire

import (
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	TypeUint24      = "uint24"
	TypeDate        = "date"
	TypeTime        = "time"
	TypeDateTime    = "datetime"
	TypeHexDate     = "hexdate"
	TypeHexTime     = "hextime"
	TypeHexDateTime = "hexdatetime"
)

// Encoder can encode an object.
type Encoder interface {
	Encode(*Writer) error
}

// Decoder can decode an object.
type Decoder interface {
	Decode(*Reader) error
}

// Encode an object to the wire.
func Encode(writer *Writer, v any) error {
	return encodeViaReflection(writer, v, encodingOptions{})
}

// Decode an object from the wire.
func Decode(reader *Reader, v any) error {
	logrus.Debugf("Decode: v: %+v", v)
	return decodeViaReflection(reader, reflect.ValueOf(v), encodingOptions{})
}

func encodeViaReflection(writer *Writer, input any, options encodingOptions) error {
	if e, ok := input.(Encoder); ok {
		logrus.Debugf("encodeViaReflection: Encoding via Encoder: %T", input)
		return e.Encode(writer)
	}

	var fieldPrefix string
	if options.Name != "" {
		fieldPrefix = "field " + options.Name + ": "
	}

	myType := reflect.TypeOf(input)
	logrus.Debugf("encodeViaReflection: Kind: %v", myType.Kind())
	switch myType.Kind() {
	case reflect.Pointer:
		logrus.Debugf("encodeViaReflection: input: %+v", input)
		if reflect.ValueOf(input).IsNil() {
			logrus.Debugf("encodeViaReflection: input is nil.")
			switch myType.Elem().Kind() {
			case reflect.Struct:
				if options.Length > 0 {
					if options.Null != nil {
						logrus.Debugf("encodeViaReflection: writing %d bytes of 0x%x.", options.Length, *options.Null)
						for i := 0; i < options.Length; i++ {
							writer.WriteUint8(*options.Null)
						}
					}
				}
				return nil
			default:
				return fmt.Errorf(fieldPrefix+"unhandled pointer type: %+v", myType.Elem().Kind())
			}
		} else {
			logrus.Debugf("encodeViaReflection: input is not nil.")
			return encodeViaReflection(writer, reflect.ValueOf(input).Elem().Interface(), options)
		}
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
			case TypeHexDate:
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Year() - 2000)))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Month())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Day())))
			case TypeHexTime:
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Hour())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Minute())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Second())))
			case TypeHexDateTime:
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Year() - 2000)))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Month())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Day())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Weekday())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Hour())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Minute())))
				writer.WriteUint8(InsaneBase10ToBase16(uint8(timeValue.Second())))
			default:
				return fmt.Errorf(fieldPrefix+"unhandled type: %s", options.Type)
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
					return err
				}
				options.Name = myField.Name
				var myFieldValue reflect.Value
				if myField.IsExported() {
					myFieldValue = myValue.Field(f)
				} else {
					myFieldValue = reflect.New(myField.Type)
				}
				err = encodeViaReflection(writer, myFieldValue.Interface(), options)
				if err != nil {
					return err
				}
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
		var bytesToWrite []byte
		if myType.Elem().Kind() == reflect.Uint8 {
			logrus.Debugf("encodeViaReflection: array value: %+v", reflect.ValueOf(input))
			if myType.Len() == 0 {
				// Cool; this is always empty.
			} else {
				bytesToWrite = reflect.ValueOf(input).Bytes()
			}
		} else {
			return fmt.Errorf(fieldPrefix+"unimplemented: encoding an array of kind %q", myType.Elem().Kind())
		}

		if options.Length > 0 {
			for len(bytesToWrite) < options.Length {
				bytesToWrite = append(bytesToWrite, 0x00)
			}
			if len(bytesToWrite) > options.Length {
				return fmt.Errorf(fieldPrefix+"invalid length: wrote %d (expected %d)", len(bytesToWrite), options.Length)
			}
		}

		logrus.Debugf("encodeViaReflection: writing array of bytes: (%d) %x", len(bytesToWrite), bytesToWrite)
		writer.WriteBytes(bytesToWrite)
	case reflect.Slice:
		var bytesToWrite []byte
		if ipValue, ok := input.(net.IP); ok {
			if ipValue.To4() == nil {
				return fmt.Errorf(fieldPrefix + "could not convert address to IPv4")
			}
			bytesToWrite = ipValue.To4()
		} else if myType.Elem().Kind() == reflect.Uint8 {
			logrus.Debugf("encodeViaReflection: slice value: %+v", reflect.ValueOf(input))
			bytesToWrite = reflect.ValueOf(input).Bytes()
		} else {
			return fmt.Errorf(fieldPrefix+"unimplemented: encoding a slice of kind %q", myType.Elem().Kind())
		}

		if options.Length > 0 {
			for len(bytesToWrite) < options.Length {
				bytesToWrite = append(bytesToWrite, 0x00)
			}
			if len(bytesToWrite) > options.Length {
				return fmt.Errorf(fieldPrefix+"invalid length: wrote %d (expected %d)", len(bytesToWrite), options.Length)
			}
		}

		logrus.Debugf("encodeViaReflection: writing slice of bytes (%d): %x", len(bytesToWrite), bytesToWrite)
		writer.WriteBytes(bytesToWrite)
	default:
		return fmt.Errorf(fieldPrefix+"unimplemented: kind: %v", myType.Kind())
	}
	return nil
}

func decodeViaReflection(reader *Reader, myValue reflect.Value, options encodingOptions) error {
	logrus.Debugf("decodeViaReflection: myValue: %+v", myValue)
	logrus.Debugf("decodeViaReflection: options: %+v", options)

	if myValue != reflect.ValueOf(nil) && myValue.CanInterface() {
		if d, ok := myValue.Interface().(Decoder); ok {
			logrus.Debugf("decodeViaReflection: Decoding via Decoder: %+v", myValue.Interface())
			return d.Decode(reader)
		}
	}

	var fieldPrefix string
	if options.Name != "" {
		fieldPrefix = "field " + options.Name + ": "
	}

	myType := myValue.Type()
	logrus.Debugf("decodeViaReflection: myType: %+v", myType)
	logrus.Debugf("decodeViaReflection: myType.Kind: %+v", myType.Kind())

	switch myType.Kind() {
	case reflect.Pointer:
		logrus.Debugf("decodeViaReflection: myType.Elem.Kind: %+v", myType.Elem().Kind())
		if myValue.IsNil() {
			switch myType.Elem().Kind() {
			case reflect.Struct:
				if options.Length == 0 {
					switch options.Type {
					case TypeDate:
						options.Length = 2
					case TypeTime:
						options.Length = 2
					case TypeDateTime:
						options.Length = 4
					}
				}
				if options.Length > 0 {
					logrus.Debugf("decodeViaReflection: reading %d bytes.", options.Length)
					subReader, err := reader.Read(options.Length)
					if err != nil {
						return fmt.Errorf(fieldPrefix+"could not read sub structure: %w", err)
					}
					logrus.Debugf("decodeViaReflection: read (%d): %x", subReader.Length(), subReader.Bytes())
					if options.Null != nil && IsAll(subReader.Bytes(), *options.Null) {
						// Leave this as null.
						return nil
					}
					newValue := reflect.New(myType.Elem())
					err = decodeViaReflection(subReader, newValue, options)
					if err != nil {
						return fmt.Errorf(fieldPrefix+"could not decode sub structure: %w", err)
					}
					logrus.Debugf("decodeViaReflection: newValue %+vx", newValue)
					myValue.Set(newValue)
				}
				return nil
			default:
				return fmt.Errorf(fieldPrefix+"unhandled pointer type: %+v", myType.Elem().Kind())
			}
		} else {
			return decodeViaReflection(reader, myValue.Elem(), options)
		}
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
					return fmt.Errorf(fieldPrefix+"could not read date: %w", err)
				}
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf(fieldPrefix+"could not set date: %w", err)
				}
			case TypeTime:
				v, err := reader.ReadTime()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read time: %w", err)
				}
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf(fieldPrefix+"could not set time: %w", err)
				}
			case TypeDateTime:
				v1, err := reader.ReadDate()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read date: %w", err)
				}
				v2, err := reader.ReadTime()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read time: %w", err)
				}
				v := MergeDateTime(v1, v2)
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf(fieldPrefix+"could not set datetime: %w", err)
				}
			case TypeHexDate:
				year, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read year: %w", err)
				}
				year = InsaneBase16ToBase10(year)
				month, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read month: %w", err)
				}
				month = InsaneBase16ToBase10(month)
				day, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read day: %w", err)
				}
				day = InsaneBase16ToBase10(day)

				v := time.Date(int(year)+2000, time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf(fieldPrefix+"could not set date: %w", err)
				}
			case TypeHexDateTime:
				year, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read year: %w", err)
				}
				year = InsaneBase16ToBase10(year)
				month, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read month: %w", err)
				}
				month = InsaneBase16ToBase10(month)
				day, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read day: %w", err)
				}
				day = InsaneBase16ToBase10(day)

				weekday, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read weekday: %w", err)
				}
				weekday = InsaneBase16ToBase10(weekday)
				_ = weekday

				hour, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read hour: %w", err)
				}
				hour = InsaneBase16ToBase10(hour)
				minute, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read minute: %w", err)
				}
				minute = InsaneBase16ToBase10(minute)
				second, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read second: %w", err)
				}
				second = InsaneBase16ToBase10(second)

				v := time.Date(int(year)+2000, time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf(fieldPrefix+"could not set date: %w", err)
				}
			case TypeHexTime:
				hour, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read hour: %w", err)
				}
				hour = InsaneBase16ToBase10(hour)
				minute, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read minute: %w", err)
				}
				minute = InsaneBase16ToBase10(minute)
				second, err := reader.ReadUint8()
				if err != nil {
					return fmt.Errorf(fieldPrefix+"could not read second: %w", err)
				}
				second = InsaneBase16ToBase10(second)

				v := time.Date(0, time.January, 1, int(hour), int(minute), int(second), 0, time.UTC)
				if myValue.CanSet() {
					myValue.Set(reflect.ValueOf(v))
				} else {
					return fmt.Errorf(fieldPrefix+"could not set date: %w", err)
				}
			default:
				return fmt.Errorf(fieldPrefix+"unhandled type: %s", options.Type)
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
				options.Name = myField.Name
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
				return fmt.Errorf(fieldPrefix+"could not read remainder: %w", err)
			}
			if myType.Len() == 0 {
				if IsAll(contents, 0) {
					logrus.Debugf("decodeViaReflection: contents is all 0.")
					contents = contents[len(contents):]
				} else {
					logrus.Debugf(fieldPrefix+"decodeViaReflection: contents is not 0: %x", contents)
				}
			}
			if myType.Len() != len(contents) {
				return fmt.Errorf(fieldPrefix+"unexpected length: %d (expected: %d)", len(contents), myType.Len())
			}
			for i := 0; i < myType.Len(); i++ {
				myValue.Index(i).Set(reflect.ValueOf(contents[i]))
			}
		} else {
			return fmt.Errorf(fieldPrefix+"unhandled array type: %v", myType.Elem().Kind())
		}
	case reflect.Slice:
		logrus.Debugf("decodeViaReflection: slice")
		if myType.Elem().Kind() == reflect.Uint8 {
			logrus.Debugf("decodeViaReflection: slice of type uint8")

			if options.Length < 0 {
				options.Length = reader.Length()
			}

			contents, err := reader.ReadBytes(options.Length)
			if err != nil {
				return fmt.Errorf(fieldPrefix+"could not read remainder: %w", err)
			}

			if options.Length != len(contents) {
				return fmt.Errorf(fieldPrefix+"unexpected length: %d (expected: %d)", len(contents), options.Length)
			}
			logrus.Debugf("decodeViaReflection: making slice of length %d", options.Length)
			newValue := reflect.MakeSlice(myType, options.Length, options.Length)
			for i := 0; i < options.Length; i++ {
				logrus.Debugf("decodeViaReflection: setting index %d to 0x%x", i, contents[i])
				newValue.Index(i).Set(reflect.ValueOf(contents[i]))
			}
			myValue.Set(newValue)
		} else {
			return fmt.Errorf(fieldPrefix+"unhandled slice type: %v", myType.Elem().Kind())
		}
	case reflect.Uint8:
		if options.Type != "" {
			return fmt.Errorf(fieldPrefix+"invalid type for uint8: %q", options.Type)
		}
		v, err := reader.ReadUint8()
		if err != nil {
			return err
		}
		logrus.Debugf("decodeViaReflection: read uint8: %v", v)
		if myValue.CanUint() {
			myValue.SetUint(uint64(v))
		} else {
			return fmt.Errorf(fieldPrefix+"could not set uint: %w", err)
		}
	case reflect.Uint16:
		if options.Type != "" {
			return fmt.Errorf(fieldPrefix+"invalid type for uint16: %q", options.Type)
		}
		v, err := reader.ReadUint16()
		if err != nil {
			return err
		}
		logrus.Debugf("decodeViaReflection: read uint16: %v", v)
		if myValue.CanUint() {
			myValue.SetUint(uint64(v))
		} else {
			return fmt.Errorf(fieldPrefix+"could not set uint: %w", err)
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
				return fmt.Errorf(fieldPrefix+"could not set uint: %w", err)
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
				return fmt.Errorf(fieldPrefix+"could not set uint: %w", err)
			}
		default:
			return fmt.Errorf(fieldPrefix+"invalid type for uint8: %q", options.Type)
		}
	default:
		return fmt.Errorf(fieldPrefix+"unhandled kind: %v", myType.Kind())
	}
	return nil
}
