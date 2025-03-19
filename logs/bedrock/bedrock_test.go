package bedrock

import (
	_ "embed"
	"strings"

	"encoding/json"
	"testing"
)

var (
	//go:embed 1.19.41.01.txt
	StaticLogFileBedrock1 string
	//go:embed 1.6.1.0.txt
	StaticLogFileBedrock2 string
	//go:embed 1.21.70-beta25.txt
	StaticLogFileBedrock3 string
)

func testPrintLog(t *testing.T, textPrint string, log *BedrockParse) {
	d, _ := json.MarshalIndent(log, "", "  ")
	t.Logf(textPrint, string(d))
}

func TestDetectAndParseBedrock(t *testing.T) {
	parsedLog, err := &BedrockParse{}, error(nil)
	if err = parsedLog.Parse(strings.NewReader(StaticLogFileBedrock1)); err != nil {
		t.Errorf("Cannot parse bedrock log 1: %s", err)
		return
	}
	testPrintLog(t, "Parsed log bedrock Static 1:\n%s", parsedLog)

	if err = parsedLog.Parse(strings.NewReader(StaticLogFileBedrock2)); err != nil {
		t.Errorf("Cannot parse bedrock log 2: %s", err)
		return
	}
	testPrintLog(t, "Parsed log bedrock Static 2:\n%s", parsedLog)

	if err = parsedLog.Parse(strings.NewReader(StaticLogFileBedrock3)); err != nil {
		t.Errorf("Cannot parse bedrock log 3: %s", err)
		return
	}
	testPrintLog(t, "Parsed log bedrock Static 3:\n%s", parsedLog)
}
