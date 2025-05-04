package allaymc

import (
	"bytes"
	"path"
	"sync"

	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/request/github"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Server version info
type Version struct {
	Version     string                   `json:"version"`      // Server version
	JavaVersion javaprebuild.JavaVersion `json:"java_version"` // Java Version, example: 21
	ServerURL   string                   `json:"download"`     // Server file to download
}

func (ver Version) SemverVersion() semver.Version { return semver.New(ver.Version) }

// Download server
func (ver Version) Dowload(path string) error {
	if ver.ServerURL != "" {
		_, err := request.SaveAs(ver.ServerURL, path, nil)
		return err
	}
	return ErrNoVersion
}

// Slice with all versions possibles
type Versions []*Version

func processVersionWorker(versions *Versions, jobs <-chan *github.Release, err *error, wg *sync.WaitGroup) {
	defer wg.Done()
	for releaseInfo := range jobs {
		for _, asset := range releaseInfo.Assets {
			if path.Ext(asset.BrowserDownloadURL) != ".jar" {
				continue
			}

			// Get server file in memory
			serverBuffer, _, err2 := request.Buffer(asset.BrowserDownloadURL, nil)
			if err2 != nil {
				*err = err2
				break
			}

			// Get server minimum java version
			jarVersion, err2 := javaprebuild.JarMajor(bytes.NewReader(serverBuffer), int64(asset.Size))
			if err2 != nil {
				*err = err2
				break
			}

			// append to versions
			*versions = append(*versions, &Version{
				Version:     releaseInfo.TagName,
				ServerURL:   asset.BrowserDownloadURL,
				JavaVersion: jarVersion,
			})
			break
		}
	}
}

// Fetch all versions from github releases
func (versions *Versions) FetchFromGithub() error {
	jobs, wg, err := make(chan *github.Release, 12), sync.WaitGroup{}, error(nil)

	// Start workers to process server versions
	for range 15 {
		wg.Add(1)
		go processVersionWorker(versions, jobs, &err, &wg)
	}

	// Make basic client to github APIs
	allayMcGitub := github.NewClient("AllayMC", "Allay", "")
	for releaseInfo, err := range allayMcGitub.ReleaseSeq() {
		if err != nil {
			close(jobs)
			return err
		} else if releaseInfo.TagName == "nightly" {
			continue
		}
		jobs <- releaseInfo
	}
	close(jobs)
	wg.Wait()

	// Sort structs
	semver.Sort(*versions)
	return err
}
