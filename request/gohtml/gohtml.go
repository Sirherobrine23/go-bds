// MIT License
//
// Thanks for Tan Nguyen (@tanqhnguyen) for module, from: https://gitlab.com/tanqhnguyen/gohtml
package gohtml

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const selectorTagName = "html"

// different types of content extractor
type contentExtractor interface {
	GetContent(selection *goquery.Selection) (string, error)
}

type textContentExtractor struct{}

func (c *textContentExtractor) GetContent(selection *goquery.Selection) (string, error) {
	return strings.TrimSpace(selection.Text()), nil
}

type tagAttributeContentExtractor struct {
	Attribute string
}

func (c *tagAttributeContentExtractor) GetContent(selection *goquery.Selection) (string, error) {
	tagValue := selection.AttrOr(c.Attribute, "")
	return strings.TrimSpace(tagValue), nil
}

// determine whether we want to go to the next DOM level or stay in the current one
// and use attribute to get the value
func getContentExtractor(sel *goquery.Selection, tagValue string) (*goquery.Selection, contentExtractor) {
	if parts := strings.Split(tagValue, ",attr"); len(parts) > 1 {
		return sel, &tagAttributeContentExtractor{Attribute: strings.TrimSpace(parts[0])}
	} else if parts := strings.Split(tagValue, ", attr"); len(parts) > 1 {
		return sel, &tagAttributeContentExtractor{Attribute: strings.TrimSpace(parts[0])}
	} else if parts := strings.Split(tagValue, "="); len(parts) > 1 {
		return sel.Find(strings.TrimSpace(parts[0])), &tagAttributeContentExtractor{Attribute: strings.TrimSpace(parts[1])}
	}

	return sel.Find(tagValue), &textContentExtractor{}
}

func recursivelyParseDoc(doc *goquery.Selection, structure any, extractor contentExtractor) error {
	if extractor == nil {
		extractor = &textContentExtractor{}
	}

	structType := reflect.TypeOf(structure)
	if structType.Kind() != reflect.Ptr {
		return fmt.Errorf("must pass a pointer")
	}
	elem := structType.Elem()

	kind := elem.Kind()

	value := reflect.ValueOf(structure)

	// parse the top level struct or nested struct/slice
	if kind == reflect.Struct {
		structurePointer := value.Elem()
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			if field.Tag == "" {
				continue
			}

			selectorTagValue := field.Tag.Get(selectorTagName)
			if selectorTagValue == "" {
				continue
			}

			fieldPointer := structurePointer.FieldByName(field.Name)
			if !fieldPointer.CanSet() {
				continue
			}

			kind := field.Type.Kind()
			targetNode, nodeContentExtractor := getContentExtractor(doc, selectorTagValue)

			switch kind {
			case reflect.Struct:
				// create a new struct pointer and recursively extract data from it
				nestedStruct := reflect.New(fieldPointer.Type()).Interface()
				if err := recursivelyParseDoc(targetNode, nestedStruct, nodeContentExtractor); err != nil {
					return err
				}
				fieldPointer.Set(reflect.ValueOf(nestedStruct).Elem())
			case reflect.Slice:
				// first get the Type of the children
				childType := fieldPointer.Type().Elem()
				// then loop through each matched elements and set the data
				htmlNodes := []*goquery.Selection{}
				targetNode.Each(func(_ int, selection *goquery.Selection) { htmlNodes = append(htmlNodes, selection) })
				for _, selection := range htmlNodes {
					element := reflect.New(childType).Interface()
					if err := recursivelyParseDoc(selection, element, nodeContentExtractor); err != nil {
						return err
					}
					fieldPointer.Set(reflect.Append(fieldPointer, reflect.ValueOf(element).Elem()))
				}
			default:
				newValue := reflect.New(fieldPointer.Type()).Interface()
				err := recursivelyParseDoc(targetNode, newValue, nodeContentExtractor)
				if err != nil {
					fmt.Printf("unable to convert value to [%s][%s]\n", fieldPointer.Type().Kind(), field.Name)
					break
				}
				fieldPointer.Set(reflect.ValueOf(newValue).Elem())
			}
		}
	}

	// handle primitive types
	htmlValue, err := extractor.GetContent(doc)
	if err != nil {
		return err
	}

	valuePtr := value.Elem()
	switch kind {
	case reflect.String:
		valuePtr.SetString(htmlValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if htmlValue == "" {
			valuePtr.SetInt(0)
			break
		}
		intValue, err := strconv.ParseInt(htmlValue, 10, 64)
		if err != nil {
			return err
		}
		valuePtr.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		if htmlValue == "" {
			valuePtr.SetFloat(0)
			break
		}
		floatValue, err := strconv.ParseFloat(htmlValue, 64)
		if err != nil {
			return err
		}
		valuePtr.SetFloat(floatValue)
	default:
		break
	}

	return nil
}

// Parse finds and extracts data from a HTML content
func Parse(content []byte, target any) error {
	return NewParse(bytes.NewReader(content), target)
}

// Parse finds and extracts data from a HTML reader stream
func NewParse(r io.Reader, target any) error {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}
	return recursivelyParseDoc(doc.Selection, target, nil)
}
