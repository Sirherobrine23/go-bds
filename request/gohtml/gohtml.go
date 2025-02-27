package gohtml

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"maps"
	"reflect"
	"slices"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

var (
	falseValues = []string{"", "false", "off", "0"}

	encodingText   = reflect.TypeFor[encoding.TextUnmarshaler]()
	encodingBinary = reflect.TypeFor[encoding.BinaryUnmarshaler]()
)

func CheckTarget(valueof reflect.Value) error {
	if valueof.IsNil() || valueof.Type().Kind() == reflect.Interface || valueof.Type().Kind() != reflect.Pointer {
		typeof := valueof.Type()
		if typeof == nil {
			return fmt.Errorf("unmarshal(nil)")
		}
		if typeof.Kind() != reflect.Pointer {
			return fmt.Errorf("unmarshal(non-pointer %s)", typeof.String())
		}
		return fmt.Errorf("unmarshal(nil %s)", typeof.String())
	}
	return nil
}

// Unmarshal finds and extracts data from a HTML content
func Unmarshal(content []byte, v any) error {
	return NewDecode(bytes.NewReader(content), v)
}

// Parse finds and extracts data from a HTML reader stream
func NewDecode(r io.Reader, v any) error {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	valueof := reflect.ValueOf(v)
	if err := CheckTarget(valueof); err != nil {
		return err
	}
	return QueryParse(doc.Selection, valueof, TextExtractor)
}

// Parse [*github.com/PuerkitoBio/goquery.Selection] to [reflect.Value]
func QueryParse(doc *goquery.Selection, target reflect.Value, extractor ContentExtractor) error {
	valueExtracted, err := extractor(doc)
	if err != nil {
		return err
	}

	// Check if implemets enconding Binary and Text
	if target.Type().Implements(encodingText) || target.Type().Implements(encodingBinary) {
		switch v := target.Interface().(type) {
		case encoding.TextUnmarshaler:
			return v.UnmarshalText([]byte(valueExtracted))
		case encoding.BinaryUnmarshaler:
			return v.UnmarshalBinary([]byte(valueExtracted))
		}
	}

	switch value := target.Elem(); value.Type().Kind() {
	case reflect.String:
		value.SetString(valueExtracted)
	case reflect.Bool:
		value.SetBool(!slices.Contains(falseValues, valueExtracted))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch valueExtracted {
		case "", "0":
			value.SetInt(0)
		default:
			intValue, err := strconv.ParseInt(valueExtracted, 10, 64)
			if err != nil {
				return err
			}
			value.SetInt(intValue)
		}
	case reflect.Float32, reflect.Float64:
		switch valueExtracted {
		case "", "0", "0.0", "0.", ".0":
			value.SetFloat(0)
		default:
			floatValue, err := strconv.ParseFloat(valueExtracted, 64)
			if err != nil {
				return err
			}
			value.SetFloat(floatValue)
		}
	case reflect.Struct:
		elem := value.Type()
		for i := range elem.NumField() {
			field := elem.Field(i)
			if selectorTagValue := field.Tag.Get(TagName); selectorTagValue == "" || selectorTagValue == "-" {
				continue
			}

			fieldTarget := value.FieldByName(field.Name)
			if !fieldTarget.CanSet() {
				continue
			}

			targetNode, nodeContentExtractor := getContentExtractor(doc, field.Tag.Get(TagName))
			switch field.Type.Kind() {
			case reflect.Struct,
				reflect.String,
				reflect.Bool,
				reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64,
				reflect.Float32,
				reflect.Float64:
				if err := QueryParse(targetNode, fieldTarget.Addr(), nodeContentExtractor); err != nil {
					return err
				}
			case reflect.Interface:
				// Ignore
			case reflect.Pointer:
				if targetNode != nil {
					if fieldTarget.IsNil() {
						fieldTarget.Set(reflect.New(fieldTarget.Type().Elem()))
					}
					if err := QueryParse(targetNode, fieldTarget, nodeContentExtractor); err != nil {
						return err
					}
				}
			case reflect.Array:
				values := slices.Collect(maps.Values(maps.Collect(targetNode.EachIter())))
				for valueIndex := range fieldTarget.Len() {
					if valueIndex < len(values) {
						if err := QueryParse(values[valueIndex], fieldTarget.Index(valueIndex).Addr(), nodeContentExtractor); err != nil {
							return err
						}
					}
				}
			case reflect.Slice:
				// first get the Type of the children
				childType := fieldTarget.Type().Elem()

				// then loop through each matched elements and set the data
				for _, selection := range targetNode.EachIter() {
					element := reflect.New(childType)
					if err := QueryParse(selection, element, nodeContentExtractor); err != nil {
						return err
					}
					fieldTarget.Set(reflect.Append(fieldTarget, element.Elem()))
				}
			}
		}
	}

	return nil
}
