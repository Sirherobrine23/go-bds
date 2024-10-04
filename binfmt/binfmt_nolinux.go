//go:build !(linux || android || windows)

// Sorry this platform not support binfmt
package binfmt

// List archs registred
func Archs() ([]*Fmt, error) { return nil, ErrNoSupportedPlatform }
