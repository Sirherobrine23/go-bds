//go:build !((aix && ppc64) || darwin || (windows && amd64) || (linux && (amd64 || arm || arm64 || riscv64 || ppc64le || s390x)) || (solaris && (amd64 || sparcv9)))

package adoptium

// Current system not supported to install prebuild java files
func InstallLatest(featVersion uint, installPath string) error {
	return ErrSystem
}
