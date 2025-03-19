package logs_test

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"testing"

	"sirherobrine23.com.br/go-bds/go-bds/logs"
	_ "sirherobrine23.com.br/go-bds/go-bds/logs/bedrock"
	_ "sirherobrine23.com.br/go-bds/go-bds/logs/java"
)

//go:embed */1.*.txt
var LogFiles embed.FS

func TestLogs(t *testing.T) {
	err := fs.WalkDir(LogFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		file, _ := LogFiles.Open(path)
		if _, err = logs.Parse(file.(io.ReadSeeker)); err != nil {
			return fmt.Errorf("cannot parse %s: %s", path, err)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
