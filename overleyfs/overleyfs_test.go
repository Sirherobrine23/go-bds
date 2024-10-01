package overleyfs

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOverlayMount(t *testing.T) {
	if !(runtime.GOOS == "linux" || runtime.GOOS == "windows") {
		t.Skip("Test not supported")
		return
	}

	root, err := os.MkdirTemp(os.TempDir(), "overlay_test_*")
	if err != nil {
		return
	}
	defer os.RemoveAll(root)
	t.Logf("Folder test %q", root)

	var fs Overlayfs

	fs.Workdir = filepath.Join(root, "workdir")
	fs.Target = filepath.Join(root, "merged")
	fs.Upper = filepath.Join(root, "upper")
	fs.Lower = []string{
		filepath.Join(root, "low1"),
		filepath.Join(root, "low2"),
	}
	for _, k := range append(fs.Lower, fs.Workdir, fs.Target, fs.Upper) {
		os.MkdirAll(k, 0600)
	}

	textExample := []byte("google is best\n")
	os.WriteFile(filepath.Join(fs.Lower[0], "test1.txt"), textExample, 0600)
	os.WriteFile(filepath.Join(fs.Upper, "test2.txt"), textExample, 0600)

	defer fs.Unmount()
	if err := fs.Mount(); err != nil {
		t.Error(err)
		return
	}

	d1, _ := os.ReadFile(filepath.Join(fs.Target, "test1.txt"))
	d2, _ := os.ReadFile(filepath.Join(fs.Target, "test2.txt"))
	for _, k := range [][]byte{d1, d2} {
		if !bytes.Equal(textExample, k) {
			t.Error("cannot check overlayfs correct work")
			break
		}
	}
	fs.Unmount()
}
