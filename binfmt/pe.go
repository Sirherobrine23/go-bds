package binfmt

import (
	"debug/pe"
	"fmt"
)

var _ Binary = (*Pe)(nil)

// Linux elf binary
type Pe pe.File

func (binPE *Pe) From() any    { return (*pe.File)(binPE) }
func (binPE *Pe) GoOs() string { return "windows" }

func (binPE *Pe) GoArch() string {
	return fmt.Sprintf("arch_%d", binPE.Machine)
}
