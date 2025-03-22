package properties

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var (
	reflectValue  = reflect.TypeFor[Node]()
	reflectText   = reflect.TypeFor[encoding.TextMarshaler]()
	reflectJSON   = reflect.TypeFor[json.Marshaler]()
	reflectUntext = reflect.TypeFor[encoding.TextUnmarshaler]()
	reflectUnjson = reflect.TypeFor[json.Unmarshaler]()
)

// Encode struct to bytes value
func Marshal(ptr any) ([]byte, error) {
	b := &bytes.Buffer{}
	return b.Bytes(), NewWrite(b).Encode(ptr)
}

type Writer struct {
	w io.Writer
}

// Create new properties write
func NewWrite(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Write properties file
func (wr *Writer) Encode(ptr any) error {
	return wr.writeStruct("", reflect.ValueOf(ptr))
}

func (wr *Writer) writeStruct(keyName string, valueOf reflect.Value) (err error) {
	if valueOf.Type().Kind() == reflect.Pointer {
		if valueOf.IsNil() {
			_, err = fmt.Fprintf(wr.w, "%s = null", keyName)
			return err
		}
		return wr.writeStruct(keyName, valueOf.Elem())
	}

	if valueOf.Type().Implements(reflectValue) || valueOf.Type().Implements(reflectText) || valueOf.Type().Implements(reflectJSON) {
		var data []byte
		switch v := valueOf.Interface().(type) {
		case nil:
			data = []byte("null")
		case encoding.TextMarshaler:
			data, err := v.MarshalText()
			if err == nil {
				_, err = fmt.Fprintf(wr.w, "%s = %s\n", keyName, data)
			}
			return err
		case json.Marshaler:
			data, err = v.MarshalJSON()
		}
		if err == nil {
			_, err = fmt.Fprintf(wr.w, "%s = %s\n", keyName, data)
		}
		return err
	}

	switch valueOf.Type().Kind() {
	case reflect.String:
		_, err = fmt.Fprintf(wr.w, "%s = %s\n", keyName, valueOf.String())
	case reflect.Bool:
		_, err = fmt.Fprintf(wr.w, "%s = %v\n", keyName, valueOf.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err = fmt.Fprintf(wr.w, "%s = %d\n", keyName, valueOf.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, err = fmt.Fprintf(wr.w, "%s = %d\n", keyName, valueOf.Uint())
	case reflect.Float32, reflect.Float64:
		_, err = fmt.Fprintf(wr.w, "%s = %f\n", keyName, valueOf.Float())
	case reflect.Interface:
		if valueOf.IsNil() {
			_, err = fmt.Fprintf(wr.w, "%s = null\n", keyName)
			return err
		} else if valueOf.IsZero() {
			data, err := json.Marshal(valueOf.Interface())
			if err == nil {
				_, err = fmt.Fprintf(wr.w, "%s = %s\n", keyName, data)
			}
			return err
		}

		if valueOf.IsValid() && !valueOf.IsNil() && valueOf.Elem().IsValid() {
			return wr.writeStruct(keyName, valueOf.Elem())
		}
	case reflect.Slice, reflect.Array:
		for keyIndex := range valueOf.Len() {
			keyNamed := fmt.Sprintf("%s[%d]", keyName, keyIndex)
			if err = wr.writeStruct(keyNamed, valueOf.Index(keyIndex)); err != nil {
				return err
			}
		}
	case reflect.Map:
		for key, value := range valueOf.Seq2() {
			var keyNamed string
			switch key.Type().Kind() {
			case reflect.String:
				keyNamed = key.String()
			case reflect.Bool:
				keyNamed = fmt.Sprintf("%v", key.Bool())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				keyNamed = fmt.Sprintf("%d", key.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				keyNamed = fmt.Sprintf("%d", key.Uint())
			case reflect.Float32, reflect.Float64:
				keyNamed = fmt.Sprintf("%f", key.Float())
			default:
				continue
			}

			if keyName != "" {
				keyNamed = fmt.Sprintf("%s.%s", keyName, keyNamed)
			}
			if err = wr.writeStruct(keyNamed, value); err != nil {
				return
			}
		}
	case reflect.Struct:
		structType := valueOf.Type()
		for keyIndex := range valueOf.NumField() {
			field, fieldType := valueOf.Field(keyIndex), structType.Field(keyIndex)
			if !fieldType.IsExported() || fieldType.Tag.Get(propertiesTagName) == "-" {
				continue
			}

			keyNamed := keyName
			tagConfig := strings.Split(fieldType.Tag.Get(propertiesTagName), ",")
			if len(tagConfig) > 0 && tagConfig[0] == "" {
				keyNamed = fieldType.Name
				if keyName != "" {
					keyNamed = fmt.Sprintf("%s.%s", keyName, fieldType.Name)
				}
			} else if len(tagConfig) > 0 && tagConfig[0] != "" {
				keyNamed = tagConfig[0]
				if keyName != "" {
					keyNamed = fmt.Sprintf("%s.%s", keyName, tagConfig[0])
				}
			}

			if err = wr.writeStruct(keyNamed, field); err != nil {
				return err
			}
		}
	}
	return
}
