package properties

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

//go:embed test.properties
var testFile string

type TestObj struct {
	Lang                 string  `properties:"language"`
	MultiDimArray        [][]int `properties:"multiDimArray"`
	AnotherMultiDimArray [][]any `properties:"anotherMultiDimArray"`
	JSON                 struct {
		Name string `json:"fieldName"`
	} `properties:"objectFromText"`

	Timed    time.Time
	Maped    map[string]string
	MapedAny map[int]any

	Object struct {
		Bool  bool    `properties:"booleanValue1"`
		Int   int     `properties:"integerNumber"`
		Float float32 `properties:"doubleNumber"`
		Man   struct {
			John string `properties:"name"`
		} `properties:"man"`
	} `properties:"object"`

	Object2 struct {
		SimpleArray [2]string           `properties:"simpleArray"`
		ObjectArray []map[string]string `properties:"objectArray"`
	} `properties:"object2"`

	Object3 struct {
		Any                any      `properties:"nullValue"`
		ArrayWithDelimeter []string `properties:"arrayWithDelimeter"`
	} `properties:"object3"`
}

func TestReader(t *testing.T) {
	r := NewParse(strings.NewReader(testFile))

	var n TestObj
	if err := r.Decode(&n); err != nil {
		t.Error(err)
		return
	}

	d, _ := json.MarshalIndent(n, "", "  ")
	t.Log(string(d))
}

func TestWrite(t *testing.T) {
	testObj := &TestObj{
		Lang:          "pt-br",
		MultiDimArray: [][]int{{0, 1, 6, 8}, nil, {2, 5, 9, 2}},
		AnotherMultiDimArray: [][]any{
			{
				map[string]string{
					"g0000001gle": "1000000000000",
					"n1":          "n2",
				},
			},
		},
		JSON: struct {
			Name string "json:\"fieldName\""
		}{
			Name: "Test name",
		},
		Timed: time.Now(),
		Maped: map[string]string{
			"go": "made by Google.Inc",
		},
		MapedAny: map[int]any{
			1: nil,
			2: "string",
			3: []int{1, 2, 3, 4},
			4: []any{1, true, false, nil, "string", 1.209},
		},
		Object: struct {
			Bool  bool    "properties:\"booleanValue1\""
			Int   int     "properties:\"integerNumber\""
			Float float32 "properties:\"doubleNumber\""
			Man   struct {
				John string "properties:\"name\""
			} "properties:\"man\""
		}{
			Bool:  true,
			Int:   1,
			Float: 23.13,
			Man: struct {
				John string "properties:\"name\""
			}{
				John: "is not john",
			},
		},
		Object3: struct {
			Any                any      "properties:\"nullValue\""
			ArrayWithDelimeter []string "properties:\"arrayWithDelimeter\""
		}{
			Any:                false,
			ArrayWithDelimeter: []string{"golang", "gopher"},
		},
	}

	buffer := &bytes.Buffer{}
	if err := NewWrite(buffer).Encode(testObj); err != nil {
		t.Error(err)
	}
	t.Log(buffer.String())
}
