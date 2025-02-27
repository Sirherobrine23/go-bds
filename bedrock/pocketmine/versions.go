package pocketmine

import (
	"encoding/json"
	"errors"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var (
	ErrInvalidFileVersions error  = errors.New("invalid versions file or url")                                            // Versions file invalid url schema
	VersionsRemote         string = "https://sirherobrine23.com.br/go-bds/Pocketmine-Cache/raw/branch/main/versions.json" // Version cache
)

type Version struct {
	Version  string    `json:"version"`
	Release  time.Time `json:"releaseTime"`
	Pharfile string    `json:"phar"`
}

// Get versions from cached versions
// remoteFileFetch set custom cache versions for load versions
func FromVersions() ([]Version, error) {
	var versions []Version
	res, err := request.Request(VersionsRemote, nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&versions); err != nil {
		return nil, err
	}

	return versions, nil
}
