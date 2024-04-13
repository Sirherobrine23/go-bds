package globals
type Version struct {
	Version string            `json:"version"`
	Targets map[string]string `json:"targets"`
}