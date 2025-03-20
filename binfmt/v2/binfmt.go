package binfmt

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"errors"
	"os"
)

var ErrNotDetect error = errors.New("cannot detect binary file")

type Binary interface {
	GoArch() string // Binary arch name with GOARCH style.
	GoOs() string   // Binary target name with GOOS style.
	From() any      // Binary from, example: [debug/elf.File], [debug/pe.File], [debug/macho.File] or nil.
}

func AsEmulator(target Binary) []string {
	switch v := target.(type) {
	case LinuxEmulator:
		return v.EmulatorCommand
	case *LinuxEmulator:
		return v.EmulatorCommand
	case LinuxMisc:
		return v.Interpreter
	case *LinuxMisc:
		return v.Interpreter
	}
	return nil
}

func Open(name string) (Binary, error) {
	localFile, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer localFile.Close()

	// Linux
	linuxEmulator, err := GetEmulatorTarget(localFile)
	if err == nil {
		return linuxEmulator, nil
	}

	// Unix ELF
	elfFile, err := elf.NewFile(localFile)
	if err == nil {
		return (*Elf)(elfFile), nil
	}

	// Windows PE
	peFile, err := pe.NewFile(localFile)
	if err == nil {
		return (*Pe)(peFile), nil
	}

	// MacOS Macho
	machoFile, err := macho.NewFile(localFile)
	if err == nil {
		return (*Macho)(machoFile), nil
	}

	return nil, ErrNotDetect
}
