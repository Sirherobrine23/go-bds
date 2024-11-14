//go:build windows

package binfmt

func Archs() ([]Binfmt, error) { return nil, ErrNoSupportedPlatform }
