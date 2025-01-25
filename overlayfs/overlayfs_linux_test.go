//go:build linux

package overlayfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxOverlayfs(t *testing.T) {
	targetPath := filepath.Join(t.TempDir(), "target")
	overlayMount := &Overlayfs{
		Target:  targetPath,
		Upper:   filepath.Join(t.TempDir(), "up"),
		Workdir: filepath.Join(t.TempDir(), "work"),
		Lower: []string{
			filepath.Join(t.TempDir(), "low1"),
			filepath.Join(t.TempDir(), "low2"),
			filepath.Join(t.TempDir(), "low3"),
			filepath.Join(t.TempDir(), "low4"),
		},
	}

	for _, folderPath := range append(overlayMount.Lower, overlayMount.Target, overlayMount.Upper, overlayMount.Workdir) {
		if err := os.Mkdir(folderPath, 0777); err != nil {
			t.Skipf("skiping, cannot make folders: %q", err.Error())
			return
		}
	}

	err := overlayMount.Mount()
	if err != nil {
		switch err {
		case fs.ErrPermission:
			t.Skip("cannot mount overlayfs, run with unshare or root user")
		default:
			t.Errorf("Cannot mount overlayfs: %q", err.Error())
		}
		return // Stop test
	}

	if err = overlayMount.Unmount(); err != nil {
		t.Errorf("unmount overlayfs: %q", err.Error())
		return
	}

	if err = overlayMount.Unmount(); err != nil {
		t.Errorf("second unmount overlayfs: %q", err.Error())
		return
	}
}
