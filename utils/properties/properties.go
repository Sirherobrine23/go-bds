// Java properties Parse and Writer
package properties

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
)

var (
	ErrLocked         error  = errors.New("cannot decode because is opened on another requests")
	propertiesTagName string = "properties"

	_ Node = (*Object)(nil)
	_ Node = (*Slice)(nil)
	_ Node = (*String)(nil)
	_ Node = (*Bool)(nil)
	_ Node = (*Float)(nil)
	_ Node = (*Int)(nil)
	_ Node = (*Null)(nil)
	_ Node = (*Json)(nil)

	kindNames = []string{
		KindNull:   "null",
		KindObject: "object",
		KindSlice:  "slice",
		KindString: "string",
		KindBool:   "bool",
		KindInt:    "int",
		KindFloat:  "float",
		KindJSON:   "json",
	}
)

func nodeStr(previus, current Node) string {
	if previus == nil && current == nil {
		return "null"
	} else if previus == nil {
		return current.Key()
	}
	return current.FullKey() + "." + current.Key()
}

// A Kind represents the specific kind of type that a [sirherobrine23.com.br/go-bds/go-bds/utils/properties.Node] represents.
type Kind int

// String returns the name of k.
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind" + strconv.Itoa(int(k))
}

const (
	KindNull Kind = iota
	KindObject
	KindSlice
	KindString
	KindBool
	KindInt
	KindFloat
	KindJSON
)

// Abstract value type
type Node interface {
	Kind() Kind                        // Value type
	Key() string                       // Key name
	FullKey() string                   // full node name, with previus names if exists
	ValueIndex(index int) (Node, bool) // Return value by index
	ValueKey(key string) (Node, bool)  // Return value by key name
	Value() any                        // Return raw value
	ValueString() string               // Return string if avaible
	ValueInt() int64                   // return int if possible
	ValueUint() uint64                 // return uint if possible
	ValueFloat() float64               // return float if possible
	ValueBool() bool                   // return bool if possible
	json.Marshaler                     // marshal json
}

// Equivalent tp map[string]any
type Object struct {
	nodeName  string
	dadNode   Node
	MapValues map[string]Node
}

func (Object) Kind() Kind                      { return KindObject }
func (m *Object) FullKey() string              { return nodeStr(m.dadNode, m) }
func (m *Object) Key() string                  { return m.nodeName }
func (m *Object) MarshalJSON() ([]byte, error) { return json.Marshal(m.MapValues) }
func (m *Object) Value() any                   { return m.MapValues }
func (m *Object) ValueInt() int64              { return int64(len(m.MapValues)) }
func (m *Object) ValueIndex(index int) (v Node, ok bool) {
	v, ok = m.MapValues[js_types.Slice[string](js_types.Maper[string, Node](m.MapValues).Keys()).At(index)]
	return
}
func (m *Object) ValueKey(key string) (v Node, ok bool) {
	if v, ok = m.MapValues[key]; !ok {
		nodePath := ParseNodePath(key)
		if sliceName, _, isSlice := nodePath.NamedSlice(); isSlice {
			v, ok = m.MapValues[sliceName]
		}
		nodePath = nodePath[1:]
		for len(nodePath) > 0 && v != nil {
			if sliceIndex, isSlice := nodePath.Slice(); isSlice {
				v, ok = v.ValueIndex(sliceIndex)
			} else if sliceName, sliceIndex, isSlice := nodePath.NamedSlice(); isSlice {
				if v, ok = v.ValueKey(sliceName); ok {
					v, ok = v.ValueIndex(sliceIndex)
				}
			}
			nodePath = nodePath[1:]
		}
	}
	return
}
func (Object) ValueBool() bool     { return false }
func (Object) ValueString() string { return "" }
func (Object) ValueUint() uint64   { return 0 }
func (Object) ValueFloat() float64 { return 0 }

type Slice struct {
	nodeName string
	dadNode  Node
	Slice    []Node
}

func (Slice) Kind() Kind                      { return KindSlice }
func (s *Slice) FullKey() string              { return nodeStr(s.dadNode, s) }
func (s *Slice) Key() string                  { return s.nodeName }
func (s *Slice) MarshalJSON() ([]byte, error) { return json.Marshal(s.Slice) }
func (s *Slice) Value() any                   { return s.Slice }
func (s *Slice) ValueInt() int64              { return int64(len(s.Slice)) }
func (Slice) ValueBool() bool                 { return false }
func (Slice) ValueString() string             { return "" }
func (Slice) ValueUint() uint64               { return 0 }
func (Slice) ValueFloat() float64             { return 0 }
func (s *Slice) ValueIndex(index int) (v Node, ok bool) {
	v = js_types.Slice[Node](s.Slice).At(index)
	ok = v != nil
	return
}

func (s *Slice) ValueKey(key string) (v Node, ok bool) {
	nodePath := ParseNodePath(key)
	if len(nodePath) > 0 {
		v = s
	}
	for len(nodePath) > 0 && v != nil {
		if sliceIndex, isSlice := nodePath.Slice(); isSlice {
			v, ok = v.ValueIndex(sliceIndex)
		} else if sliceName, sliceIndex, isSlice := nodePath.NamedSlice(); isSlice {
			if v, ok = v.ValueKey(sliceName); ok {
				v, ok = v.ValueIndex(sliceIndex)
			}
		} else {
			v, ok = v.ValueKey(nodePath[0])
		}
		nodePath = nodePath[1:]
	}
	return
}

type String struct {
	nodeName string
	dadNode  Node
	String   string
}

func (String) Kind() Kind                        { return KindString }
func (s *String) FullKey() string                { return nodeStr(s.dadNode, s) }
func (s *String) Key() string                    { return s.nodeName }
func (s *String) MarshalJSON() ([]byte, error)   { return json.Marshal(s.String) }
func (s *String) Value() any                     { return s.String }
func (s *String) ValueString() string            { return s.String }
func (s *String) ValueInt() int64                { return int64(len(s.String)) }
func (String) ValueKey(string) (Node, bool)      { return nil, false }
func (String) ValueIndex(index int) (Node, bool) { return nil, false }
func (String) ValueUint() uint64                 { return 0 }
func (String) ValueFloat() float64               { return 0 }
func (String) ValueBool() bool                   { return false }
func (s *String) SplitString(sep string) []Node {
	stringsPifs := []Node{}
	for str := range strings.SplitSeq(s.String, sep) {
		stringsPifs = append(stringsPifs, &String{String: str, dadNode: s})
	}
	return stringsPifs
}

type Bool struct {
	nodeName string
	dadNode  Node
	v        bool
}

func (Bool) Kind() Kind                        { return KindBool }
func (b *Bool) FullKey() string                { return nodeStr(b.dadNode, b) }
func (b *Bool) Key() string                    { return b.nodeName }
func (b *Bool) MarshalJSON() ([]byte, error)   { return json.Marshal(b.v) }
func (b *Bool) Value() any                     { return b.v }
func (b *Bool) ValueBool() bool                { return b.v }
func (Bool) ValueKey(string) (Node, bool)      { return nil, false }
func (Bool) ValueIndex(index int) (Node, bool) { return nil, false }
func (Bool) ValueString() string               { return "" }
func (Bool) ValueInt() int64                   { return 0 }
func (Bool) ValueUint() uint64                 { return 0 }
func (Bool) ValueFloat() float64               { return 0 }

type Float struct {
	nodeName string
	dadNode  Node
	f        float64
}

func (Float) Kind() Kind                        { return KindFloat }
func (f *Float) FullKey() string                { return nodeStr(f.dadNode, f) }
func (f *Float) Key() string                    { return f.nodeName }
func (f *Float) MarshalJSON() ([]byte, error)   { return fmt.Appendf(nil, "%f", f.f), nil }
func (f *Float) Value() any                     { return f.f }
func (f *Float) ValueFloat() float64            { return f.f }
func (Float) ValueKey(string) (Node, bool)      { return nil, false }
func (Float) ValueIndex(index int) (Node, bool) { return nil, false }
func (Float) ValueString() string               { return "" }
func (Float) ValueInt() int64                   { return 0 }
func (Float) ValueUint() uint64                 { return 0 }
func (Float) ValueBool() bool                   { return false }

type Int struct {
	nodeName string
	dadNode  Node
	v        int64
}

func (Int) Kind() Kind                        { return KindInt }
func (Int) ValueKey(string) (Node, bool)      { return nil, false }
func (Int) ValueIndex(index int) (Node, bool) { return nil, false }
func (i *Int) FullKey() string                { return nodeStr(i.dadNode, i) }
func (i *Int) Key() string                    { return i.nodeName }
func (i *Int) MarshalJSON() ([]byte, error)   { return fmt.Appendf(nil, "%d", i.v), nil }
func (i *Int) Value() any                     { return i.v }
func (i *Int) ValueInt() int64                { return i.v }
func (i *Int) ValueUint() uint64              { return uint64(i.v) }
func (Int) ValueString() string               { return "" }
func (Int) ValueFloat() float64               { return 0 }
func (Int) ValueBool() bool                   { return false }

type Null struct {
	nodeName string
	dadNode  Node
}

func (Null) Kind() Kind                        { return KindNull }
func (n *Null) FullKey() string                { return nodeStr(n.dadNode, n) }
func (n *Null) Key() string                    { return n.nodeName }
func (Null) ValueKey(string) (Node, bool)      { return nil, false }
func (Null) ValueIndex(index int) (Node, bool) { return nil, false }
func (Null) MarshalJSON() ([]byte, error)      { return json.Marshal(nil) }
func (Null) Value() any                        { return nil }
func (Null) ValueString() string               { return "" }
func (Null) ValueInt() int64                   { return 0 }
func (Null) ValueUint() uint64                 { return 0 }
func (Null) ValueFloat() float64               { return 0 }
func (Null) ValueBool() bool                   { return false }

// Implements generic JSON parse from string values
type Json struct {
	nodeName string
	dadNode  Node
	JS       json.RawMessage
}

func (Json) Kind() Kind          { return KindJSON }
func (js *Json) FullKey() string { return nodeStr(js.dadNode, js) }
func (js *Json) Key() string     { return js.nodeName }

func (js *Json) ValueString() string          { return string(js.JS) }
func (js *Json) MarshalJSON() ([]byte, error) { return js.JS, nil }
func (js *Json) Value() (data any) {
	json.Unmarshal(js.JS, &data)
	return
}

func (js *Json) ValueInt() (data int64) {
	json.Unmarshal(js.JS, &data)
	return
}
func (js *Json) ValueUint() (data uint64) {
	json.Unmarshal(js.JS, &data)
	return
}
func (js *Json) ValueFloat() (data float64) {
	json.Unmarshal(js.JS, &data)
	return
}
func (js *Json) ValueBool() (data bool) {
	json.Unmarshal(js.JS, &data)
	return
}

func (js *Json) ValueKey(key string) (Node, bool) {
	var data any
	if json.Unmarshal(js.JS, &data) == nil {
		value, ok := ParseValue(js, data)
		if ok {
			return value.ValueKey(key)
		}
	}
	return nil, false
}

func (js *Json) ValueIndex(index int) (Node, bool) {
	var data any
	if json.Unmarshal(js.JS, &data) == nil {
		value, ok := ParseValue(js, data)
		if ok {
			return value.ValueIndex(index)
		}
	}
	return nil, false
}
