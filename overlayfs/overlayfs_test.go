//go:build linux || (windows && (amd64 || 386 || arm64))

package overlayfs

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestOverlayMount(t *testing.T) {
	root := filepath.Join(t.TempDir(), "overlayfs")
	if err := os.MkdirAll(root, 0666); err != nil {
		return
	}
	// defer os.RemoveAll(root)
	t.Logf("Root folder %q", root)
	var overlayFS Overlayfs
	overlayFS.Workdir = filepath.Join(root, "workdir")
	overlayFS.Target = filepath.Join(root, "merged")
	overlayFS.Upper = filepath.Join(root, "upper")
	overlayFS.Lower = []string{
		filepath.Join(root, "low1"),
		filepath.Join(root, "low2"),
	}
	for _, k := range append(overlayFS.Lower, overlayFS.Workdir, overlayFS.Target, overlayFS.Upper) {
		os.MkdirAll(k, 0600)
	}

	config, _ := json.MarshalIndent(overlayFS, "", "  ")
	t.Log(string(config))

	textExample := []byte("google is best\n")
	os.WriteFile(filepath.Join(overlayFS.Lower[0], "test1.txt"), textExample, 0600)
	os.WriteFile(filepath.Join(overlayFS.Upper, "test2.txt"), textExample, 0600)

	defer overlayFS.Unmount()
	if err := overlayFS.Mount(); err != nil {
		switch os.Getenv("CI") {
		case "1", "true":
			t.Skipf("Skiped because CI Fail mount Overlayfs or not compatible, error: %s", err.Error())
		default:
			t.Error(err)
		}
		return
	}

	<-time.After(time.Minute * 2)
	entrys, err := os.ReadDir(overlayFS.Target)
	if err != nil {
		t.Error("cannot check overlayfs correct work")
		return
	}
	entrys = slices.DeleteFunc(entrys, func(s fs.DirEntry) bool { return s.Name() == "_obgmgrproj.guid" })
	if len(entrys) != 2 {
		files := ""
		for _, i := range entrys {
			if files += ", " + i.Name(); strings.HasPrefix(files, ", ") {
				files = files[2:]
			}
		}
		t.Errorf("Invalid entrys files count, files count %d: %q", len(entrys), files)
		return
	}

	d1, _ := os.ReadFile(filepath.Join(overlayFS.Target, "test1.txt"))
	d2, _ := os.ReadFile(filepath.Join(overlayFS.Target, "test2.txt"))
	for _, k := range [][]byte{d1, d2} {
		if !bytes.Equal(textExample, k) {
			t.Error("cannot check overlayfs correct work")
			break
		}
	}

	if err := os.WriteFile(filepath.Join(overlayFS.Target, "test3.txt"), textExample, 0600); err != nil {
		t.Error(err)
		return
	}

	d3, err := os.ReadFile(filepath.Join(overlayFS.Upper, "test3.txt"))
	if err != nil {
		t.Error(err)
		return
	} else if !bytes.Equal(textExample, d3) {
		t.Error("cannot check overlayfs write to Top layer")
		return
	}

	if err := overlayFS.Unmount(); err != nil {
		t.Logf("Unmount error: %s", err.Error())
	}
}
