//go:build !((linux || windows || darwin || solaris || aix) && (amd64 || 386 || ppc64 || ppc64le || s390x || arm64 || arm || sparcv9 || riscv64))

package downloads

import "errors"

func InstallLatest(featVersion uint, path string) error {
	return errors.New("not implemented")
}
