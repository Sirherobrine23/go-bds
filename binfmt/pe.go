package binfmt

import (
	"debug/pe"
	"fmt"
)

var _ Binary = (*Pe)(nil)

// Linux elf binary
type Pe pe.File

func (*Pe) GoVariant() string  { return "" }
func (binPE *Pe) Close() error { return (*pe.File)(binPE).Close() }
func (binPE *Pe) From() any    { return (*pe.File)(binPE) }
func (binPE *Pe) GoOs() string { return "windows" }

func (binPE *Pe) GoArch() string {
	switch binPE.Machine {
	case pe.IMAGE_FILE_MACHINE_AMD64:
		return "amd64"
	case pe.IMAGE_FILE_MACHINE_I386:
		return "386"
	case pe.IMAGE_FILE_MACHINE_ARMNT:
		return "arm"
	case pe.IMAGE_FILE_MACHINE_ARM64:
		return "arm64"
	case pe.IMAGE_FILE_MACHINE_IA64:
		return "ia64"
	default:
		return fmt.Sprintf("arch_%d", binPE.Machine)
	}
}
