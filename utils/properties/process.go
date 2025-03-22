package properties

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Process key to sliced named with indexer
//
// example:
//
//	"root.node.key"  => ["root", "node", "key"]
//	"root[index][0]" => ["root[index]", "[0]"]
//	"root[0][0]"     => ["root[0]", "[0]"]
type NodeKey []string

func (n NodeKey) MarshalText() (text []byte, err error) { return []byte(n.String()), nil }
func (n NodeKey) String() string {
	node := ""
	for _, value := range n {
		switch {
		case node == "":
			node = value
		case strings.HasPrefix(value, "["):
			node += value
		default:
			node += "." + value
		}
	}
	return node
}
func (n NodeKey) IsSlice() bool {
	return len(n) >= 1 && strings.HasPrefix(n[0], "[") && strings.HasSuffix(n[0], "]")
}

// [Index] => Value
func (n NodeKey) Slice() (int, bool) {
	if len(n) >= 1 && strings.HasPrefix(n[0], "[") && strings.HasSuffix(n[0], "]") {
		index, _ := strconv.Atoi(strings.Trim(n[0], "[]"))
		return index, true
	}
	return -1, false
}

func (n NodeKey) IsNamedSlice() bool {
	return len(n) >= 1 && !strings.HasPrefix(n[0], "[") && strings.Contains(n[0], "[")
}

// Object[Name][Index] => Value
func (n NodeKey) NamedSlice() (Name string, index int, ok bool) {
	if n.IsNamedSlice() {
		bef, aft, _ := strings.Cut(n[0], "[")
		index, _ := strconv.Atoi(strings.TrimRight(aft, "]"))
		return bef, index, true
	}
	return "", -1, false
}

// Resolve properties key to Slice string
//
// Example:
//
//	"java.go[1].value" => ["java", "go[1]", "value"]
//	"java[1][2].value" => ["java[1]", "[2]", "value"]
func ParseNodePath(path string) NodeKey {
	if strings.HasPrefix(path, ".[") {
		path = "root" + path[1:]
	}
	path = strings.TrimPrefix(path, ".")

	nodeBase := []string{}
	for node := range strings.SplitSeq(path, ".") {
		for {
			if !strings.Contains(node, "][") {
				break
			}
			bef, aft, _ := strings.Cut(node, "][")
			nodeBase = append(nodeBase, bef+"]")
			node = "[" + aft
		}
		if node != "" {
			nodeBase = append(nodeBase, node)
		}
	}
	return nodeBase
}

// Set valid Value to Value
func ProcessPrimiteValue(previus Node, name, value string) Node {
	switch strings.ToLower(value) {
	case "null":
		return &Null{nodeName: name, dadNode: previus}
	case "":
		return &String{nodeName: name, dadNode: previus, String: value}
	case "true", "false":
		return &Bool{nodeName: name, dadNode: previus, v: strings.ToLower(value) == "true"}
	}

	if intV, err := strconv.ParseInt(value, 10, 0); err == nil {
		return &Int{nodeName: name, dadNode: previus, v: intV}
	}

	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return &Float{nodeName: name, dadNode: previus, f: v}
	}

	var js json.RawMessage
	if json.Unmarshal([]byte(value), &js) == nil {
		return &Json{nodeName: name, dadNode: previus, JS: js}
	}

	return &String{nodeName: name, dadNode: previus, String: value}
}

// Insert value to struct apropriety
func ProcessStruct(previus Node, name NodeKey, value string) error {
	if previus == nil || len(name) == 0 {
		return nil
	}
	for len(name) > 1 {
		nodeName := name[0]
		switch v := previus.(type) {
		case *Object:
			if sliceName, sliceIndex, isSlice := name.NamedSlice(); isSlice {
				if v.MapValues[sliceName] == nil {
					v.MapValues[sliceName] = &Slice{nodeName: sliceName, dadNode: v, Slice: []Node{}}
				}

				slice, ok := v.MapValues[sliceName].(*Slice)
				if !ok {
					return nil
				}

				if sliceIndex := sliceIndex + 1; sliceIndex > len(slice.Slice) {
					slice.Slice = append(slice.Slice, make([]Node, sliceIndex-len(slice.Slice))...)
				}

				previus = slice
				name = name[1:]

				// If is not slice insert new object if not exists
				if !name.IsSlice() {
					if slice.Slice[sliceIndex] == nil {
						newObj := &Object{nodeName: sliceName, dadNode: slice, MapValues: map[string]Node{}}
						slice.Slice[sliceIndex] = newObj
					}
					previus = slice.Slice[sliceIndex]
					continue
				} else if slice.Slice[sliceIndex] == nil {
					newSlice := &Slice{nodeName: sliceName, dadNode: slice, Slice: []Node{}}
					slice.Slice[sliceIndex] = newSlice
				}

				previus = slice.Slice[sliceIndex]
				continue
			}

			// Create object if not exists with Name from name[0]
			if v.MapValues[nodeName] == nil {
				v.MapValues[nodeName] = &Object{
					nodeName:  nodeName,
					dadNode:   v,
					MapValues: map[string]Node{},
				}
			}

			previus = v.MapValues[nodeName]
			name = name[1:]
			continue
		case *Slice:
			// If is Named slice, create object before insert new slice into
			if sliceName, sliceIndex, isSlice := name.NamedSlice(); isSlice {
				if sliceIndex := sliceIndex + 1; sliceIndex > len(v.Slice) {
					v.Slice = append(v.Slice, make([]Node, sliceIndex-len(v.Slice))...)
				}
				if v.Slice[sliceIndex] == nil {
					v.Slice[sliceIndex] = &Object{nodeName: sliceName, dadNode: v, MapValues: map[string]Node{}}
				}
				previus = v.Slice[sliceIndex]
				name = name[1:]
				continue
			}

			// If is slice index append value to current
			if sliceIndex, isSlice := name.Slice(); isSlice {
				if sliceIndex := sliceIndex + 1; sliceIndex > len(v.Slice) {
					v.Slice = append(v.Slice, make([]Node, sliceIndex-len(v.Slice))...)
				}
				if v.Slice[sliceIndex] == nil {
					v.Slice[sliceIndex] = &Slice{nodeName: fmt.Sprintf("[%d]", sliceIndex), dadNode: v, Slice: []Node{}}
				}
				previus = v.Slice[sliceIndex]
				name = name[1:]
				continue
			}

			target := &Object{nodeName: nodeName, dadNode: v, MapValues: map[string]Node{}}
			v.Slice = append(v.Slice, target)
			previus = target
			name = name[1:]
			continue
		default:
			return nil
		}
	}

	if len(name) == 1 {
		switch previus := previus.(type) {
		case *Slice:
			if sliceIndex, isSlice := name.Slice(); isSlice {
				if sliceIndex := sliceIndex + 1; sliceIndex > len(previus.Slice) {
					previus.Slice = append(previus.Slice, make([]Node, sliceIndex-len(previus.Slice))...)
				}

				previus.Slice[sliceIndex] = ProcessPrimiteValue(previus, name[0], value)
				return nil
			}

			return fmt.Errorf("cannot set %T %q: %q => %q, require slice index", previus, previus.FullKey(), name[0], value)
		case *Object:
			sliceName, sliceIndex, isSlice := name.NamedSlice()
			if isSlice {
				if previus.MapValues[sliceName] == nil {
					previus.MapValues[sliceName] = &Slice{
						nodeName: sliceName,
						dadNode:  previus,
						Slice:    make([]Node, sliceIndex+1),
					}
				}

				previus, ok := previus.MapValues[sliceName].(*Slice)
				if !ok {
					return nil
				}

				if sliceIndex := sliceIndex + 1; sliceIndex > len(previus.Slice) {
					previus.Slice = append(previus.Slice, make([]Node, sliceIndex-len(previus.Slice))...)
				}

				previus.Slice[sliceIndex] = ProcessPrimiteValue(previus, name[0], value)
				return nil
			}

			previus.MapValues[name[0]] = ProcessPrimiteValue(previus, name[0], value)
			return nil
		}
		return nil
	}

	// Process last child
	return nil
}

func ParseValue(previus Node, value any) (Node, bool) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return &Int{dadNode: previus, v: reflect.ValueOf(v).Int()}, true
	case uint, uint8, uint16, uint32, uint64:
		return &Int{dadNode: previus, v: int64(reflect.ValueOf(v).Uint())}, true
	case bool:
		return &Bool{dadNode: previus, v: v}, true
	case string:
		return &String{dadNode: previus, String: v}, true
	case []any:
		newSlice := &Slice{dadNode: previus, Slice: []Node{}}
		for value := range v {
			newValue, ok := ParseValue(newSlice, value)
			if ok {
				newSlice.Slice = append(newSlice.Slice, newValue)
			}
		}
		return newSlice, true
	case map[any]any:
		newObj := &Object{dadNode: previus, MapValues: map[string]Node{}}
		for key, value := range v {
			keyName := ""
			switch v := key.(type) {
			case string:
				keyName = v
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				keyName = fmt.Sprintf("%d", v)
			case float32, float64:
				keyName = fmt.Sprintf("%f", v)
			case bool:
				keyName = fmt.Sprintf("%v", v)
			default:
				continue
			}

			newValue, ok := ParseValue(newObj, value)
			if ok {
				newObj.MapValues[keyName] = newValue
			}
		}
		return newObj, true
	}
	return nil, false
}
