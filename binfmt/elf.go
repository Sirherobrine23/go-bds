package binfmt

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"strings"
)

var _ Binary = (*Elf)(nil)

// Linux elf binary
type Elf elf.File

func (binElf *Elf) From() any    { return (*elf.File)(binElf) }
func (binElf *Elf) Close() error { return (*elf.File)(binElf).Close() }
func (binElf *Elf) GoOs() string {
	switch (*elf.File)(binElf).OSABI {
	case elf.ELFOSABI_LINUX, elf.ELFOSABI_NONE:
		return "linux"
	case elf.ELFOSABI_HURD:
		return "hurd"
	case elf.ELFOSABI_NETBSD:
		return "netbsd"
	case elf.ELFOSABI_FREEBSD:
		return "freebsd"
	case elf.ELFOSABI_OPENBSD:
		return "openbsd"
	case elf.ELFOSABI_SOLARIS:
		return "solaris"
	case elf.ELFOSABI_AIX:
		return "aix"
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
	case elf.EM_860, elf.EM_960:
		return "386"
	case elf.EM_PPC64:
		if binElf.FileHeader.ByteOrder == binary.LittleEndian {
			return "ppc64le"
		}
		return "ppc64"
	case elf.EM_S390:
		switch binElf.Class {
		case elf.ELFCLASS64:
			return "s390x"
		default:
			return "s390"
		}
	case elf.EM_RISCV:
		sect := (*elf.File)(binElf).Section(".riscv.attributes")
		if sect != nil {
			data, err := sect.Data()
			if err == nil {
				if idx := bytes.Index(data, []byte("\x05rv")); idx != -1 {
					end := bytes.IndexByte(data[idx:], 0x00)
					if end == -1 {
						end = len(data)
					} else {
						end += idx
					}
					isa := string(data[idx:end])
					if strings.HasPrefix(isa, "\x05rv32") {
						return "riscv"
					} else if strings.HasPrefix(isa, "\x05rv64") {
						return "riscv64"
					} else if strings.HasPrefix(isa, "\x05rv128") {
						return "riscv128"
					}
				}
			}
		}

		switch binElf.Class {
		case elf.ELFCLASS32:
			return "riscv"
		case elf.ELFCLASS64:
			return "riscv64"
		}

		return "riscv"
	case elf.EM_MIPS:
		var isa string
		if binElf.Class == elf.ELFCLASS64 {
			isa = "64"
		}

		// mips, mips64, mips64le, mipsle
		switch binElf.FileHeader.ByteOrder {
		case binary.BigEndian:
			return "misp" + isa
		case binary.LittleEndian:
			return fmt.Sprintf("mips%sle", isa)
		default:
			return "mips"
		}
	default:
		return strings.ToLower(strings.TrimPrefix(binElf.Machine.String(), "EM_"))
	}
}

func (binElf *Elf) GoVariant() string {
	switch binElf.Machine {
	case elf.EM_ARM: // arm/v4, arm/v5, arm/v6, arm/v7
		armAttributesSection := (*elf.File)(binElf).Section(".ARM.attributes")
		if armAttributesSection != nil {
			data, err := armAttributesSection.Data()
			if err != nil {
				return ""
			}

			// Validate that the section starts with the format version 'A'
			if data[0] != 'A' {
				return ""
			}

			parseULEB128 := func(data []byte) (result uint32, n int) {
				for shift := uint(0); n < len(data); shift += 7 {
					b := data[n]
					result |= uint32(b&0x7f) << shift
					n++
					if (b & 0x80) == 0 {
						return result, n
					}
				}
				return 0, 0
			}

			i := 1
			for i < len(data) {
				// Read subsection length (4 bytes LE)
				if i+4 > len(data) {
					break
				}
				length := int(binary.LittleEndian.Uint32(data[i : i+4]))
				if length == 0 || i+length > len(data) {
					break
				}

				// Read vendor name (null-terminated string)
				j := i + 4
				for j < i+length && data[j] != 0 {
					j++
				}
				if j >= len(data) {
					break
				}
				vendor := string(data[i+4 : j])
				if vendor != "aeabi" {
					i += length
					continue
				}

				// Attributes start after vendor name + null byte
				k := j + 1
				for k < i+length {
					tag := data[k]
					k++
					// Read ULEB128 tag value
					val, n := parseULEB128(data[k:])
					if n == 0 {
						break
					}
					k += n

					if tag == 6 { // Tag_CPU_arch
						// armel => 4
						// armhf => 10
						switch val {
						case 10:
							return "v7"
						case 4:
							return "v6"
						default:
							return "v5"
						}
					}
				}
				i += length
			}
		}
	}
	return ""
}
