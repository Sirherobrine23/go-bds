// Java properties
package properties

import (
	"encoding"
	"encoding/json"
	"errors"
)

type ValueType uint

const (
	ValueTypeUnknown ValueType = iota
	ValueTypeNull
	ValueTypeArray
	ValueTypeObject
	ValueTypeString
	ValueTypeBool
	ValueTypeFloat
)

var (
	propertiesTagName = "properties"

	ErrLocked error = errors.New("cannot decode because is opened on another requests")

	_ Value = (*Null)(nil)
	_ Value = (*Array)(nil)
)

// Abstract value type
type Value interface {
	Type() ValueType       // Return value type
	KeyName() string       // Key name
	Value() any            // Return value
	encoding.TextMarshaler // return text of object, if value implemets [encoding.TextMarshaler] return this value
	json.Marshaler         // marshal json strings
}

type Null struct {
	Name string // Key name if presents
}

func (Null) Type() ValueType                       { return ValueTypeNull }
func (Null) Value() any                            { return nil }
func (Null) MarshalText() (text []byte, err error) { return []byte("null"), nil }
func (Null) MarshalJSON() ([]byte, error)          { return json.Marshal(nil) }
func (n Null) KeyName() string                     { return n.Name }

type Array struct {
	Name   string
	Values []any
}

func (Array) Type() ValueType                         { return ValueTypeArray }
func (n Array) Value() any                            { return n.Values[:] }
func (n Array) KeyName() string                       { return n.Name }
func (n Array) MarshalJSON() ([]byte, error)          { return json.Marshal(n.Values) }
func (n Array) MarshalText() (text []byte, err error) { return []byte("Array"), nil }
