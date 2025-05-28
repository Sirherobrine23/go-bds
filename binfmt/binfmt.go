// Abstract [debug/elf], [debug/male], [debug/pe] to geric interface, plus a touch to Linux Binfmt-Misc, Box64 and QEMU
package binfmt

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"errors"
	"io"
	"os"
)

var ErrNotDetect error = errors.New("cannot detect binary file")

type Binary interface {
	From() any    // Binary from, example: [debug/elf.File], [debug/pe.File], [debug/macho.File] or nil.
	Close() error // If interface implemantes close

	GoArch() string    // Binary arch name with GOARCH style.
	GoVariant() string // Binary arch variant.
	GoOs() string      // Binary target name with GOOS style.
}

// Get emulator args to run, if not is emulator return nil.
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

// Return string from Binary with variant,
// <os>/<arch>(/<variant>) same docker style
func String(target Binary) string {
	str := target.GoOs() + "/" + target.GoArch()
	if v := target.GoVariant(); v != "" {
		str += "/" + v
	}
	return str
}

// Open file and return [Binary]
func Open(name string) (Binary, error) {
	localFile, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return GetBinary(localFile)
}

// Return valid binary from [io.ReaderAt]
func GetBinary(r io.ReaderAt) (Binary, error) {
	// Linux
	linuxEmulator, err := GetEmulatorTarget(r)
	if err == nil {
		return linuxEmulator, nil
	}

	// Unix ELF
	elfFile, err := elf.NewFile(r)
	if err == nil {
		return (*Elf)(elfFile), nil
	}

	// Windows PE
	peFile, err := pe.NewFile(r)
	if err == nil {
		return (*Pe)(peFile), nil
	}

	// MacOS Macho
	machoFile, err := macho.NewFile(r)
	if err == nil {
		return (*Macho)(machoFile), nil
	}

	return nil, ErrNotDetect
}
