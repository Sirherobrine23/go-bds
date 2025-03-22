package properties

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"
)

//go:embed test.properties
var testFile string

type TestObj struct {
	John                 string    `properties:"object.man.name"`
	Lang                 string    `properties:"language"`
	Bool                 bool      `properties:"object.booleanValue1"`
	Int                  int       `properties:"object.integerNumber"`
	Float                float32   `properties:"object.doubleNumber"`
	Any                  any       `properties:"object3.nullValue"`
	Array                [2]string `properties:"object2.simpleArray"`
	MultiDimArray        [][]int   `properties:"multiDimArray"`
	AnotherMultiDimArray [][]any   `properties:"anotherMultiDimArray"`
	ArrayWithDelimeter   []string  `properties:"object3.arrayWithDelimeter"`
	JSON                 struct {
		Name string `json:"fieldName"`
	} `properties:"objectFromText"`
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

func TestLines(t *testing.T) {
	r := NewParse(strings.NewReader(testFile))
	lines, err := r.processLines()
	if err != nil {
		t.Error(err)
		return
	}
	d, _ := json.MarshalIndent(lines, "", "  ")
	t.Log(string(d))
}
