package pocketmine

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"time"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

var ErrInvalidFileVersions error = errors.New("invalid versions file or url") // Versions file invalid url schema
const (
	VersionsRemote string = "https://sirherobrine23.org/Minecraft-Server/Pocketmine-Cache/raw/branch/main/versions.json"
)

type Version struct {
	Version  string    `json:"version"`
	Release  time.Time `json:"releaseTime"`
	Pharfile string    `json:"phar"`
}

// Get versions from cached versions
// remoteFileFetch set custom cache versions for load versions
func FromVersions(remoteFileFetch ...string) ([]Version, error) {
	fileFatch := VersionsRemote
	if len(remoteFileFetch) == 1 && len(remoteFileFetch[0]) > 2 {
		fileFatch = remoteFileFetch[0]
	}

	file, err := url.Parse(fileFatch)
	if err != nil {
		return []Version{}, err
	}

	versions := []Version{}
	if file.Scheme == "http" || file.Scheme == "https" {
		res, err := request.Request(request.RequestOptions{Method: "GET", HttpError: true, Url: fileFatch})
		if err != nil {
			return versions, err
		}

		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&versions); err != nil {
			return versions, err
		}
	} else if file.Scheme == "file" {
		osFile, err := os.Open(file.Path)
		if err != nil {
			return versions, err
		}

		defer osFile.Close()
		if err = json.NewDecoder(osFile).Decode(&versions); err != nil {
			return versions, err
		}
	} else {
		return versions, ErrInvalidFileVersions
	}

	return versions, nil
}