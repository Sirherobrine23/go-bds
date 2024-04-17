package globals

import (
	"errors"
	"time"
)

var (
	DefaultTime      time.Duration = time.Hour * 5
	ErrNoSupportArch error         = errors.New("platform arch not supported")
	ErrNoSupportOs   error         = errors.New("platform os not supported")
	ErrNoTargets     error         = errors.New("current os not supported by dist, install this version in host manualy")
)

type VersionBundle struct {
	FileUrl  string `json:"fileUrl"`
	Checksum string `json:"Checksum"`
}
type Version map[int]VersionBundle
