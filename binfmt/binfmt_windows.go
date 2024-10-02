//go:build windows

package binfmt

import "runtime"

func Archs() ([]*Fmt, error) {
	switch runtime.GOARCH {
	case "amd64", "arm":
		return []*Fmt{
			{
				AutoEmulate: true,
				Arch: "386",
				Magic: []byte{0x4d, 0x5a},
			},
		}, nil
	case "arm64":
		return []*Fmt{
			{
				AutoEmulate: true,
				Arch: "386",
				Magic: []byte{0x4d, 0x5a},
			},
			{
				AutoEmulate: true,
				Arch: "x86_64",
				Magic: []byte{0x4d, 0x5a},
			},
			{
				AutoEmulate: true,
				Arch: "arm",
				Magic: []byte{0x4d, 0x5a},
			},
		}, nil
	}
	return nil, ErrNoSupportedPlatform
}