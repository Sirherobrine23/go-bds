package globals

import "time"

var DefaultTime time.Duration = time.Hour * 5

type Version struct {
	Version string            `json:"version"`
	Targets map[string]string `json:"targets"`
}
