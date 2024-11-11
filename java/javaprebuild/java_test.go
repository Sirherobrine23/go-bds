package javaprebuild

import (
	"path/filepath"
	"testing"
)

func TestInstallJava(t *testing.T) {
	if err := InstallLatest(21, filepath.Join(t.TempDir(), "javatest")); err != nil && err != ErrSystem {
		t.Error(err)
		return
	}
}
