package binfmt

import (
	"os/exec"
	"testing"
)

func TestBinfmt(t *testing.T) {
	goPath, _ := exec.LookPath("go")
	_, err := ResolveBinfmt(goPath)
	if err == ErrCannotFind || err == ErrNoSupportedPlatform {
		return
	} else if err != nil {
		t.Fatal(err)
		return
	}
	t.Fatal("Process running in emulator")
}
