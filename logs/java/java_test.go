package java

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"
)

func testPrintLog(t *testing.T, textPrint string, log *JavaParse) {
	d, _ := json.MarshalIndent(log, "", "  ")
	t.Logf(textPrint, string(d))
}

func TestDetectAndParseJava(t *testing.T) {
	parsedLog, err := &JavaParse{}, error(nil)
	if err = parsedLog.Parse(strings.NewReader(StaticLogFileJava1)); err != nil {
		t.Errorf("Cannot parse java log 1: %s", err)
		return
	}
	testPrintLog(t, "Parsed log java Static 1:\n%s", parsedLog)

	if err = parsedLog.Parse(strings.NewReader(StaticLogFileJava2)); err != nil {
		t.Errorf("Cannot parse java log 2: %s", err)
		return
	}
	testPrintLog(t, "Parsed log java Static 2:\n%s", parsedLog)
}

var (
	//go:embed 1.21.4.txt
	StaticLogFileJava1 string
	//go:embed 1.7.10.txt
	StaticLogFileJava2 string
)
