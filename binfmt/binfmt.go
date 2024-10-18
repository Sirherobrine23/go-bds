package binfmt

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"strings"
)

var (
	ErrNoSupportedPlatform error = errors.New("current platform not support binfmt")
	ErrCannotFind          error = errors.New("cannot find Fmt")
)

type Fmt struct {
	AutoEmulate bool   // System auto select emulater
	Interpreter string // binary path
	Arch        string // Binary target
	Offset      int
	Magic       []byte
	Mask        []byte
	Flags       []string
}

func FindByPlatform(target string) (bool, error) {
	archs, err := Archs()
	if err != nil {
		return false, err
	}
	if strings.Contains(target, "/") {
		target = strings.Split(target, "/")[1]
	}
	return slices.ContainsFunc(archs, func(entry *Fmt) bool {
		if entry.Arch == target {
			return true
		}
		switch target {
		case "amd64", "x86_64", "x64":
			return entry.Arch == "x86_64"
		case "i386", "i286", "x86":
			return entry.Arch == "i286" || entry.Arch =="i386" || entry.Arch == "x86"
		case "aarch64", "arm64":
			return entry.Arch == "aarch64" || entry.Arch == "arm64"
		case "arm", "armhf", "armel":
			return entry.Arch == "arm" || entry.Arch == "armhf" || entry.Arch == "armel"
		}
		return entry.Arch == target
	}), nil
}

// Check binary is required emulate software
func CheckEmulate(binPath string) (bool, error) {
	fmt, err := GetBinfmtEmulater(binPath)
	if err != nil {
		return false, err
	}
	return !fmt.AutoEmulate, nil
}

// Check if binary contains in fist offset+Magic
func GetBinfmtEmulater(binPath string) (*Fmt, error) {
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
		fistBytes := make([]byte, binFmt.Offset+len(binFmt.Magic))
		if _, err := bin.Read(fistBytes); err != nil {
			return nil, err
		} else if bytes.Contains(fistBytes, binFmt.Magic) {
			return binFmt, nil
		}
		bin.Seek(0, 0)
		continue
	}

	return nil, ErrCannotFind
}
