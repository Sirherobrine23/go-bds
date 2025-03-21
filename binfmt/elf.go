package binfmt

import (
	"debug/elf"
	"strings"
)

var _ Binary = (*Elf)(nil)

// Linux elf binary
type Elf elf.File

func (binElf *Elf) From() any { return (*elf.File)(binElf) }
func (binElf *Elf) GoOs() string {
	switch v := strings.ToLower(strings.TrimPrefix((*elf.File)(binElf).OSABI.String(), "ELFOSABI_")); v {
	case "linux", "hurd", "netbsd", "freebsd", "openbsd", "solaris", "aix":
		return v
	default:
		return "none"
	}
}

func (binElf *Elf) GoArch() string {
	switch binElf.Machine {
	case elf.EM_AARCH64:
		return "arm64"
	case elf.EM_X86_64:
		return "amd64"
	default:
		return strings.ToLower(strings.TrimPrefix(binElf.Machine.String(), "EM_"))
	}
}
