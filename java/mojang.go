package java

import (
	"sync"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

func MultiRun[E any](input []E, slice int, fn func(E, int) error) error {
	errCh := make(chan error, slice)
	defer close(errCh)

	for len(input) > 0 {
		c1 := make([]E, slice)
		n := copy(c1, input)
		input, c1 = input[n:], c1[0:n]

		w1 := sync.WaitGroup{}
		for i, d := range c1 {
			w1.Add(1)
			go func() {
				defer w1.Done()
				if err := fn(d, i); err != nil {
					errCh <- err
				}
			}()
		}
		w1.Wait()
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

func ListMojang() (Versions, error) {
	type PistonVersion struct {
		Url     string `json:"url"`
		Release string `json:"type"`
	}

	type pistonInfo struct {
		Latest   map[string]string `json:"latest"`
		Versions []PistonVersion   `json:"versions"`
	}

	type mojangPistonPackage struct {
		Version        string `json:"id"`
		Type           string `json:"type"`
		FilesDownloads map[string]struct {
			FileSize int64  `json:"size"`
			FileUrl  string `json:"url"`
			Sha1     string `json:"sha1"`
		} `json:"downloads"`
		Java struct {
			VersionMajor uint   `json:"majorVersion"`
			Component    string `json:"component"`
		} `json:"javaVersion"`
	}

	data, _, err := request.JSON[pistonInfo]("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", nil)
	if err != nil {
		return nil, err
	}

	Version := Versions{}
	err = MultiRun(data.Versions, 4, func(version PistonVersion, _ int) error {
		if version.Release != "release" {
			return nil
		}

		releaseInfo, _, err := request.JSON[mojangPistonPackage](version.Url, nil)
		if err != nil {
			return err
		} else if serverFile, ok := releaseInfo.FilesDownloads["server"]; ok {
			if releaseInfo.Java.VersionMajor == 0 {
				releaseInfo.Java.VersionMajor = 8
			}
			Version = append(Version, GenericVersion{
				Version:    releaseInfo.Version,
				JVMVersion: releaseInfo.Java.VersionMajor,
				URLs: []struct {
					Name string
					URL  string
				}{{Name: ServerName, URL: serverFile.FileUrl}},
			})
		}
		return nil
	})
	return Version, err
}
