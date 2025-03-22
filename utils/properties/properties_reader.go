package properties

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// An InvalidUnmarshalError describes an invalid argument passed to [Unmarshal].
// (The argument to [Unmarshal] must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "properties: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "properties: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "properties: Unmarshal(nil " + e.Type.String() + ")"
}

type Reader struct {
	r         io.Reader
	locked    bool
	rootValue *Object
	err       error
}

// Unmarshal data to target point
func Unmarshal(b []byte, ptr any) error {
	return NewParse(bytes.NewReader(b)).Decode(ptr)
}

// Creta new reader to process properties lines
func NewParse(r io.Reader) *Reader {
	return &Reader{
		r:      r,
		locked: false,
	}
}

// Scan lines if not started, before return Root Object with lines
func (r *Reader) Values() (Node, error) {
	if err := r.scanLines(); err != nil && err != ErrLocked {
		return nil, err
	}
	return r.rootValue, nil
}

// Scan lines from reader, if ared avaible lines append to struct
func (r *Reader) Decode(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	if err := r.scanLines(); err != nil && err != ErrLocked {
		return err
	}

	return r.unmarshalPtr(r.rootValue, rv.Elem())
}

func (r *Reader) scanLines() error {
	if r.locked {
		return ErrLocked
	}
	r.locked = true
	r.rootValue = &Object{nodeName: "", dadNode: nil, MapValues: map[string]Node{}}

	endHeader := 0
	bufferLocal := make([]byte, 4096)
	for {
		if r.err == io.EOF {
			r.err = nil
			break
		} else if r.err == nil {
			endHeader, r.err = r.r.Read(bufferLocal[endHeader:])
			if r.err != nil && r.err != io.EOF {
				return r.err
			}
		} else {
			return r.err
		}

		restartHead := true
		textToProcess := string(bufferLocal[:endHeader])
		for strings.Contains(textToProcess, "\\\n") {
			before, after, _ := strings.Cut(textToProcess, "\\\n")
			if r.err == nil && (after == "" || strings.HasSuffix(textToProcess, "\\\n")) {
				endHeader = bytes.LastIndex(bufferLocal[:endHeader-2], []byte("\n"))
				restartHead = false
				break
			}
			textToProcess = before + strings.TrimLeftFunc(after, unicode.IsSpace)
		}

		if restartHead {
			endHeader = 0
		}

		textToProcess = strings.TrimSpace(textToProcess)
		for textToProcess != "" {
			if textToProcess[0] == '\n' {
				if textToProcess = textToProcess[1:]; textToProcess == "" {
					break
				}
			}

			// Check if is comment
			for len(textToProcess) > 0 && (textToProcess[0] == '#' || textToProcess[0] == '!') {
				switch findBreak := strings.Index(textToProcess, "\n"); findBreak {
				case -1:
					textToProcess = ""
				default:
					textToProcess = textToProcess[findBreak+1:]
				}
			}

			// if text is blank break
			if textToProcess == "" {
				break
			}

			keyToProcess := textToProcess
			if keyToProcess, textToProcess, _ = strings.Cut(textToProcess, "\n"); keyToProcess == "" {
				continue
			}

			delimiter := func() int {
				skipNextRune := false
				for _, r := range "=:\t\f " {
					for lineIndex := range keyToProcess {
						if keyToProcess[lineIndex] == '\\' {
							skipNextRune = true
							continue
						} else if skipNextRune {
							skipNextRune = false
							continue
						} else if keyToProcess[lineIndex] == byte(r) {
							return lineIndex
						}
					}
				}
				return -1
			}()

			if delimiter == -1 {
				if err := ProcessStruct(r.rootValue, ParseNodePath(keyToProcess), ""); err != nil {
					return err
				}
				continue
			}

			key, value := strings.TrimRightFunc(strings.ReplaceAll(keyToProcess[:delimiter], "\\", ""), unicode.IsSpace), strings.TrimLeftFunc(keyToProcess[delimiter+1:], unicode.IsSpace)
			if err := ProcessStruct(r.rootValue, ParseNodePath(key), value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r Reader) unmarshalPtr(values Node, ptr reflect.Value) (err error) {
	defer func(err *error) {
		if err2 := recover(); err2 != nil {
			*err = fmt.Errorf("panic: %s", err2)
		}
	}(&err)

	ptrType := ptr.Type()
	if ptrType.Implements(reflectValue) {
		ptr.Set(reflect.ValueOf(values))
		return nil
	} else if ptrType.Implements(reflectUntext) {
		return ptr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(values.ValueString()))
	} else if ptrType.Implements(reflectUnjson) {
		data, err := values.MarshalJSON()
		if err != nil {
			return err
		}
		return ptr.Interface().(json.Unmarshaler).UnmarshalJSON(data)
	}

	switch v := ptrType.Kind(); v {
	case reflect.String:
		ptr.SetString(values.ValueString())
	case reflect.Bool:
		ptr.SetBool(values.ValueBool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ptr.SetInt(values.ValueInt())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ptr.SetUint(values.ValueUint())
	case reflect.Float32, reflect.Float64:
		ptr.SetFloat(values.ValueFloat())
	case reflect.Pointer:
		if ptr.IsNil() {
			if values.Kind() == KindNull {
				return nil
			}
			ptr.Set(reflect.New(ptrType))
		}
		return r.unmarshalPtr(values, ptr.Elem())
	case reflect.Interface:
		if ptr.IsZero() {
			data, err := values.MarshalJSON()
			if err != nil {
				return err
			}
			return json.Unmarshal(data, ptr.Addr().Interface())
		}
	case reflect.Array:
		elent := reflect.New(ptrType).Elem()
		for index := range elent.Len() {
			values, ok := values.ValueIndex(index)
			if !ok {
				return nil
			}
			if err := r.unmarshalPtr(values, elent.Index(index)); err != nil {
				return err
			}
		}
		ptr.Set(elent)
	case reflect.Slice:
		elent := reflect.MakeSlice(ptrType, 0, 0)
		switch values.Kind() {
		case KindString:
			for _, valueTarget := range values.(*String).SplitString(",") {
				target := reflect.New(ptrType.Elem()).Elem()
				if err := r.unmarshalPtr(valueTarget, target); err != nil {
					return err
				}
				elent = reflect.Append(elent, target)
			}
		case KindJSON:
			data, err := values.MarshalJSON()
			if err != nil {
				return err
			}
			return json.Unmarshal(data, ptr.Addr().Interface())
		default:
			for index := range int(values.ValueInt()) {
				values, ok := values.ValueIndex(index)
				if !ok {
					return nil
				}
				target := reflect.New(ptrType.Elem()).Elem()
				if err := r.unmarshalPtr(values, target); err != nil {
					return err
				}
				elent = reflect.Append(elent, target)
			}
		}
		ptr.Set(elent)
	case reflect.Map:
		ptr.Set(reflect.MakeMap(ptrType))
		keyType, valueType := ptrType.Key(), ptrType.Elem()
		switch values.Kind() {
		case KindSlice:
			switch valueType.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				for index, value := range values.(*Slice).Slice {
					target := reflect.New(valueType).Elem()
					if err := r.unmarshalPtr(value, target); err != nil {
						return err
					}
					ptr.SetMapIndex(reflect.ValueOf(index), target)
				}
			}
		case KindObject:
			switch keyType.Kind() {
			case reflect.String:
				for key, value := range values.(*Object).MapValues {
					target := reflect.New(valueType).Elem()
					if err := r.unmarshalPtr(value, target); err != nil {
						return err
					}
					ptr.SetMapIndex(reflect.ValueOf(key), target)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				for key, value := range values.(*Object).MapValues {
					if valueIndex, err := strconv.ParseInt(strings.Trim(key, "[]"), 10, 0); err == nil {
						target := reflect.New(valueType).Elem()
						if err := r.unmarshalPtr(value, target); err != nil {
							return err
						}
						ptr.SetMapIndex(reflect.ValueOf(valueIndex), target)
					}
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				for key, value := range values.(*Object).MapValues {
					if valueIndex, err := strconv.ParseUint(strings.Trim(key, "[]"), 10, 0); err == nil {
						target := reflect.New(valueType).Elem()
						if err := r.unmarshalPtr(value, target); err != nil {
							return err
						}
						ptr.SetMapIndex(reflect.ValueOf(valueIndex), target)
					}
				}
			case reflect.Float32, reflect.Float64:
				for key, value := range values.(*Object).MapValues {
					if valueIndex, err := strconv.ParseFloat(strings.Trim(key, "[]"), 64); err == nil {
						target := reflect.New(valueType).Elem()
						if err := r.unmarshalPtr(value, target); err != nil {
							return err
						}
						ptr.SetMapIndex(reflect.ValueOf(valueIndex), target)
					}
				}
			}
		}
	case reflect.Struct:
		switch values.Kind() {
		case KindJSON, KindString:
			data, err := values.MarshalJSON()
			if err != nil {
				return err
			}
			return json.Unmarshal(data, ptr.Addr().Interface())
		case KindObject:
			for fieldIndex := range ptrType.NumField() {
				fieldType, field := ptrType.Field(fieldIndex), ptr.Field(fieldIndex)
				if !fieldType.IsExported() || fieldType.Tag.Get(propertiesTagName) == "" || fieldType.Tag.Get(propertiesTagName) == "-" {
					continue
				}

				fieldName := fieldType.Tag.Get(propertiesTagName)
				if fieldName == "" {
					fieldName = fieldType.Name
				}
				fieldName = strings.Split(fieldName, ",")[0]

				if value, ok := values.ValueKey(fieldName); ok && value.Kind() != KindNull {
					if err := r.unmarshalPtr(value, field); err != nil {
						return err
					}
					continue
				}
			}
		}

	}

	return nil
}
