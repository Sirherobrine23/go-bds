package allaymc

import (
	"bytes"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"

	"sirherobrine23.com.br/go-bds/go-bds/semver"
	"sirherobrine23.com.br/go-bds/go-bds/utils/javaprebuild"
	"sirherobrine23.com.br/go-bds/request/github"
	"sirherobrine23.com.br/go-bds/request/v2"
)

// Server version info
type Version struct {
	Version     string                   `json:"version"`      // Server version
	ServerURL   string                   `json:"download"`     // Server file to download
	JavaVersion javaprebuild.JavaVersion `json:"java_version"` // Java Version, example: 21
}

func (ver Version) SemverVersion() *semver.Version { return semver.New(ver.Version) }

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
	semver.SortStruct(*versions)
	return err
}

type githubPagination struct {
	NextPage, PrevPage, FirstPage, LastPage int
	NextPageToken, Cursor, Before, After    string
}

func newPaginator(he *http.Response) *githubPagination {
	r := &githubPagination{}
	if links, ok := he.Header["Link"]; ok && len(links) > 0 {
		for link := range strings.SplitSeq(links[0], ",") {
			segments := strings.Split(strings.TrimSpace(link), ";")
			if len(segments) < 2 {
				continue
			}
			if !strings.HasPrefix(segments[0], "<") || !strings.HasSuffix(segments[0], ">") {
				continue
			}
			url, err := url.Parse(segments[0][1 : len(segments[0])-1])
			if err != nil {
				continue
			}
			q := url.Query()
			if cursor := q.Get("cursor"); cursor != "" {
				for _, segment := range segments[1:] {
					switch strings.TrimSpace(segment) {
					case `rel="next"`:
						r.Cursor = cursor
					}
				}
				continue
			}
			page := q.Get("page")
			since := q.Get("since")
			before := q.Get("before")
			after := q.Get("after")
			if page == "" && before == "" && after == "" && since == "" {
				continue
			}
			if since != "" && page == "" {
				page = since
			}
			for _, segment := range segments[1:] {
				switch strings.TrimSpace(segment) {
				case `rel="next"`:
					if r.NextPage, err = strconv.Atoi(page); err != nil {
						r.NextPageToken = page
					}
					r.After = after
				case `rel="prev"`:
					r.PrevPage, _ = strconv.Atoi(page)
					r.Before = before
				case `rel="first"`:
					r.FirstPage, _ = strconv.Atoi(page)
				case `rel="last"`:
					r.LastPage, _ = strconv.Atoi(page)
				}
			}
		}
	}
	return r
}
