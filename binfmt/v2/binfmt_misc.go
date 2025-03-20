package binfmt

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/utils/slice"
)

var miscIgnoreFiles = []string{"register", "status"}

const (
	miscDir string = "/proc/sys/fs/binfmt_misc"

	FlagUnknown   LinuxMiscFlag = iota // Unknown flag
	FlagFixBinary                      // Fix binary path
	FlagCredentials
	FlagOpenBinary
	FlagPreserveargv
)

type LinuxMiscFlag uint

func (flag LinuxMiscFlag) String() string {
	switch flag {
	case FlagFixBinary:
		return "fix-binary"
	case FlagCredentials:
		return "credentials"
	case FlagOpenBinary:
		return "open-binary"
	case FlagPreserveargv:
		return "preserve-argv[0]"
	default:
		fl, newFL := "FCOP", []byte{}
		if flag&FlagFixBinary != 0 {
			newFL = append(newFL, fl[0])
		}
		if flag&FlagCredentials != 0 {
			newFL = append(newFL, fl[1])
		}
		if flag&FlagOpenBinary != 0 {
			newFL = append(newFL, fl[2])
		}
		if flag&FlagPreserveargv != 0 {
			newFL = append(newFL, fl[3])
		}
		if v := string(newFL); v != "" {
			return v
		}
		return "Unknown"
	}
}
func (flag LinuxMiscFlag) Flags() []string {
	var flags []string

	attemps := 5
	for flag > 0 && attemps > 0 {
		switch {
		case FlagFixBinary&flag != 0:
			flags = append(flags, FlagFixBinary.String())
			flag &= FlagFixBinary
		case FlagCredentials&flag != 0:
			flags = append(flags, FlagCredentials.String())
			flag &= FlagCredentials
		case FlagOpenBinary&flag != 0:
			flags = append(flags, FlagOpenBinary.String())
			flag &= FlagOpenBinary
		case FlagPreserveargv&flag != 0:
			flags = append(flags, FlagPreserveargv.String())
			flag &= FlagPreserveargv
		default:
			attemps--
		}
	}
	return flags
}

type LinuxMisc struct {
	Enabled     bool     // System auto select emulater
	Interpreter []string // binary command
	Arch        string   // Binary target
	OS          string
	Offset      int
	Flags       LinuxMiscFlag
	Magic       []byte
	Mask        []byte
}

func (m LinuxMisc) From() any      { return m }
func (m LinuxMisc) GoArch() string { return m.Arch }
func (m LinuxMisc) GoOs() string {
	if m.OS == "" {
		return "linux"
	}
	return m.OS
}

func LinuxMiscs() ([]Binary, error) {
	files, err := os.ReadDir(miscDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = ErrNotDetect
		}
		return nil, err
	}

	interpreters := []Binary{}
	for _, fileEntry := range slice.Slice[os.DirEntry](files).Filter(func(input os.DirEntry) bool { return !slices.Contains(miscIgnoreFiles, input.Name()) }) {
		miscBuff, err := os.ReadFile(filepath.Join(miscDir, fileEntry.Name()))
		if err != nil {
			return nil, err
		}

		interpreter := &LinuxMisc{
			Arch:        strings.TrimPrefix(fileEntry.Name(), "qemu-"),
			Magic:       []byte{},
			Mask:        []byte{},
			Interpreter: []string{},
		}

		for line := range strings.SplitSeq(strings.TrimSpace(string(miscBuff)), "\n") {
			fields := strings.Fields(line)
			switch fields[0] {
			case "enabled":
				interpreter.Enabled = true
			case "interpreter":
				interpreter.Interpreter = fields[1:]
				if slice.Slice[string](interpreter.Interpreter).Find(func(input string) bool { return strings.Contains(input, "/wine") }) != "" {
					interpreter.OS = "windows"
				}
			case "flags:":
				interpreter.Flags = 0
				for letter := range strings.SplitSeq(fields[1], "") {
					switch strings.ToUpper(letter) {
					case "F", "f":
						interpreter.Flags |= FlagFixBinary
					case "C", "c":
						interpreter.Flags |= FlagCredentials
					case "O", "o":
						interpreter.Flags |= FlagOpenBinary
					case "P", "p":
						interpreter.Flags |= FlagPreserveargv
					}
				}
			case "offset":
				if _, err := fmt.Sscan(strings.TrimSpace(fields[1]), &interpreter.Offset); err != nil {
					return nil, err
				}
			case "magic":
				if interpreter.Magic, err = hex.DecodeString(fields[1]); err != nil {
					return nil, err
				}
			case "mask":
				if interpreter.Mask, err = hex.DecodeString(fields[1]); err != nil {
					return nil, err
				}
			}
		}

		interpreters = append(interpreters, interpreter)
	}

	return interpreters, nil
}
