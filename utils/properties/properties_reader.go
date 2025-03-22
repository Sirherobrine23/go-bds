package properties

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"sirherobrine23.com.br/go-bds/go-bds/internal/scanner"
	"sirherobrine23.com.br/go-bds/go-bds/utils/mapers"
)

type Decoder struct {
	r      io.Reader
	locked bool
}

func Unmarshal(b []byte, ptr any) error {
	return NewParse(bytes.NewReader(b)).Decode(ptr)
}

func NewParse(r io.Reader) *Decoder {
	return &Decoder{
		r:      r,
		locked: false,
	}
}

func removeComments(buff []byte) []byte {
	lineString := strings.Split(string(buff), "\n")
	lines := []string{}
	for currentLine := 0; currentLine < len(lineString); currentLine++ {
		line := lineString[currentLine]
		if lline := strings.TrimSpace(line); lline == "" || lline[0] == '#' || lline[0] == '!' {
			continue
		}

	repeatAppend:
		lines = append(lines, line)
		if strings.HasSuffix(line, "\\\n") {
			if currentLine+1 > len(lineString) {
				break
			}
			currentLine++
			line = lineString[currentLine]
			goto repeatAppend
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

func indexAny(s, chars string) int {
	for _, c := range chars {
		if strings.ContainsRune(s, c) {
			return strings.IndexRune(s, c)
		}
	}
	return -1
}

func (r *Decoder) processLines() (map[string]string, error) {
	if r.locked {
		return nil, ErrLocked
	}
	r.locked = true

	scan := scanner.NewScannerSplit(r.r, func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, []byte("\n")); i != -1 {
			for bytes.HasSuffix(data[:i+1], []byte("\\\n")) {
				j := bytes.Index(data[i+1:], []byte("\n"))
				if j == -1 {
					// Request more data.
					return 0, nil, nil
				}
				i += j
			}
			// We have a full newline-terminated line.
			return i + 1, removeComments(data[:i+1]), nil
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), removeComments(data), nil
		}

		// Request more data.
		return 0, nil, nil
	})

	linesProcessed := map[string]string{}
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" {
			continue
		}

		delimiter := indexAny(line, "=:\t\f ")
		if delimiter == -1 {
			linesProcessed[line] = ""
			continue
		}

		key, value := strings.TrimRightFunc(line[:delimiter], unicode.IsSpace), strings.TrimLeftFunc(line[delimiter+1:], unicode.IsSpace)
		if strings.ContainsFunc(key, unicode.IsSpace) {
			delimiter := strings.IndexFunc(line, unicode.IsSpace)
			key, value = strings.TrimRightFunc(line[:delimiter], unicode.IsSpace), strings.TrimLeftFunc(line[delimiter+1:], unicode.IsSpace)
		}

		replaceLinesBreak := strings.Split(value, "\\\n")
		for index := 1; index < len(replaceLinesBreak); index++ {
			replaceLinesBreak[index] = strings.TrimLeftFunc(replaceLinesBreak[index], unicode.IsSpace)
		}
		value = strings.Join(replaceLinesBreak, "\\\n")

		linesProcessed[key] = strings.ReplaceAll(value, "\\\n", "")
	}
	return linesProcessed, scan.Err()
}

func (r *Decoder) Decode(ptr any) error {
	point := reflect.ValueOf(ptr)
	if point.IsNil() || point.Type().Kind() != reflect.Pointer {
		if point.IsNil() {
			return fmt.Errorf("cannot decode struct because is nil")
		}
		return fmt.Errorf("cannot decode struct because is not pointter")
	}
	keysMap, err := r.processLines()
	if err != nil {
		return err
	}
	return appendToStruct(keysMap, strTag{}, point.Elem())
}

type strTag struct {
	Name, ExtraName      string
	IsZero, IsOmmitempty bool
}

func (tag strTag) String() string {
	if tag.ExtraName != "" {
		if tag.ExtraName[len(tag.ExtraName)-1] != '.' && len(tag.Name) > 0 && tag.Name[0] != '[' {
			return tag.ExtraName + "." + tag.Name
		}
		return tag.ExtraName + tag.Name
	}
	return tag.Name
}

func (tag strTag) Chield(tagStr string) strTag {
	return strTag{
		ExtraName:    tag.String(),
		Name:         tagStr,
		IsZero:       tag.IsZero,
		IsOmmitempty: tag.IsOmmitempty,
	}
}

func processTag(tag string) strTag {
	n := mapers.Slice[string](strings.Split(tag, ","))
	if len(n) == 0 {
		return strTag{}
	}

	IsZero, IsOmmitempty := n.At(1) == "omitzero", n.At(1) == "ommitempty"
	return strTag{Name: n.At(0), IsZero: IsZero, IsOmmitempty: IsOmmitempty}
}

func appendToStruct(data mapers.Maper[string, string], keyName strTag, ptr reflect.Value) (err error) {
	// defer func(err *error) {
	// 	if err2 := recover(); err2 != nil {
	// 		*err = fmt.Errorf("panic: %s", err2)
	// 	}
	// }(&err)

	switch v := ptr.Kind(); v {
	case reflect.Pointer:
		if ptr.IsNil() {
			ptr.Set(reflect.New(ptr.Type()))
		}
		return appendToStruct(data, keyName, ptr.Elem())
	case reflect.Interface:
		if ptr.IsZero() {
			ptr.Set(reflect.ValueOf(data))
			return nil
		}
	case reflect.String:
		ptr.SetString(data.Get(keyName.String()))
	case reflect.Float32, reflect.Float64:
		bit := 64
		if v == reflect.Float32 {
			bit = 32
		}

		value := data.Get(keyName.String())
		if value == "null" || value == "" {
			return nil
		}

		v, err := strconv.ParseFloat(value, bit)
		if err != nil {
			return err
		}
		ptr.SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := data.Get(keyName.String())
		if value == "null" || value == "" {
			return nil
		}
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		ptr.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value := data.Get(keyName.String())
		if value == "null" || value == "" {
			return nil
		}
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		ptr.SetUint(v)
	case reflect.Struct:
		str := ptr.Type()
		for fieldIndex := range str.NumField() {
			strType, field := str.Field(fieldIndex), ptr.Field(fieldIndex)
			if !strType.IsExported() || strType.Tag.Get(propertiesTagName) == "" || strType.Tag.Get(propertiesTagName) == "-" {
				continue
			}

			tag := processTag(strType.Tag.Get(propertiesTagName))
			value, ok := data[tag.Name]
			if (!ok || value == "") && (tag.IsOmmitempty || tag.IsZero) || value == "null" {
				continue
			}

			isJSON := func(str string) bool {
				var js json.RawMessage
				return json.Unmarshal([]byte(str), &js) == nil
			}

			// Check if is json and process
			if isJSON(value) {
				if field.Kind() == reflect.Pointer {
					if err = json.Unmarshal([]byte(value), field.Interface()); err == nil {
						continue
					}
				} else if field.CanAddr() {
					if err = json.Unmarshal([]byte(value), field.Addr().Interface()); err == nil {
						continue
					}
				}
			}

			if err = appendToStruct(data, keyName.Chield(tag.String()), field); err != nil {
				return
			}
		}
	case reflect.Slice, reflect.Array:
		var sliceTarget reflect.Value
		if v == reflect.Slice {
			sliceTarget = reflect.MakeSlice(ptr.Type(), 0, 0)
		} else {
			sliceTarget = reflect.New(ptr.Type()).Elem()
		}

		if data.HasKey(keyName.String()) {
			fields := strings.Split(data.Get(keyName.String()), ",")
			if v == reflect.Slice {
				sliceTarget = reflect.MakeSlice(ptr.Type(), len(fields), len(fields))
			} else {
				sliceTarget = reflect.New(ptr.Type()).Elem()
				fields = fields[:min(len(fields), ptr.Type().Len())]
			}

			for index := range fields {
				indexName := fmt.Sprintf("[%d]", index)
				if err := appendToStruct(map[string]string{indexName: strings.TrimSpace(fields[index])}, strTag{Name: indexName, IsZero: true, IsOmmitempty: true}, sliceTarget.Index(index)); err != nil {
					return err
				}
			}

			ptr.Set(sliceTarget)
			return nil
		}

		for valueKey := range data.Filter(func(key string) bool { return strings.HasPrefix(key, keyName.String()) }) {
			fixedName := valueKey[len(keyName.ExtraName):]
			fist, last := strings.Index(fixedName, "[")+1, strings.Index(fixedName, "]")
			index, err := strconv.Atoi(fixedName[fist:last])
			if err != nil {
				return err
			}

			if n := index + 1; sliceTarget.Len() < n {
				if v == reflect.Array {
					continue
				}
				sliceTarget = reflect.AppendSlice(sliceTarget, reflect.MakeSlice(ptr.Type(), n-sliceTarget.Len(), n-sliceTarget.Len()))
			}

			if err := appendToStruct(data, keyName.Chield(fmt.Sprintf("[%s]", fixedName[fist:last])), sliceTarget.Index(index)); err != nil {
				return err
			}
		}

		ptr.Set(sliceTarget)
	}

	return nil
}
