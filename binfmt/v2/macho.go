package binfmt

import (
	"debug/macho"
)

var _ Binary = (*Macho)(nil)

// Linux elf binary
type Macho macho.File

func (binMacho *Macho) From() any    { return (*macho.File)(binMacho) }
func (binMacho *Macho) GoOs() string { return "darwin" }

func (binMacho *Macho) GoArch() string {
	switch binMacho.Cpu {
	case macho.Cpu386:
		return "386"
	case macho.CpuAmd64:
		return "amd64"
	case macho.CpuArm:
		return "arm"
	case macho.CpuArm64:
		return "arm64"
	case macho.CpuPpc:
		return "ppc"
	case macho.CpuPpc64:
		return "ppc64"
	default:
		return ""
	}
}
