package binfmt

import (
	"errors"
	"io"
	"os"
	"slices"
	"strings"
)

var (
	ErrNoSupportedPlatform error = errors.New("current platform not support binfmt")
	ErrCannotFind          error = errors.New("cannot find Fmt")
)

type Binfmt interface {
	String() string
	Arch() string                // Target arch
	Sys() string                 // Target system
	SystemSelect() bool          // System select Interpreter automatic
	ProgramArgs() []string       // Program args
	Check(r io.ReadSeeker) error // Check if is compatible file stream
}

func Target(target string) (bool, error) {
	archs, err := Archs()
	if err != nil {
		return false, err
	} else if strings.Contains(target, "/") {
		target = strings.Split(target, "/")[1]
	}
	return slices.ContainsFunc(archs, func(entry Binfmt) bool { return entry.Arch() == target }), nil
}

// Check binary is required emulate software
func RequireEmulate(binPath string) (bool, error) {
	fmt, err := ResolveBinfmt(binPath)
	if err != nil {
		return false, err
	}
	return !fmt.SystemSelect(), nil
}

// Check if binary contains in fist offset+Magic
func ResolveBinfmt(binPath string) (Binfmt, error) {
	archs, err := Archs()
	if err != nil {
		return nil, err
	}
	bin, err := os.OpenFile(binPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer bin.Close()

	for _, binFmt := range archs {
		if err := binFmt.Check(bin); err != nil {
			if err == ErrNoSupportedPlatform {
				continue
			}
			return nil, err
		}
		bin.Seek(0, 0)
		continue
	}

	return nil, ErrCannotFind
}
