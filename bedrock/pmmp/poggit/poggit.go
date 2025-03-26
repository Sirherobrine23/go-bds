// Poggit client
package poggit

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"time"

	"sirherobrine23.com.br/go-bds/request/v2"
)

var DefaultAPI, _ = url.Parse("https://poggit.pmmp.io") // Default Poggit client

// Base struct to Poggit client
type Poggit struct {
	Host *url.URL // Poggit host, default is https://poggit.pmmp.io
}

// Pocketmine Info
type Pmapi struct {
	ID           int      `json:"id"`
	Incompatible bool     `json:"incompatible"`
	Indev        bool     `json:"indev"`
	Supported    bool     `json:"supported"`
	Description  []string `json:"description"`
	PHP          []string `json:"php"`
	Phar         struct {
		Default string `json:"default"`
	} `json:"phar"`
}

// PMMP Pocketmine versions
type Pmapis map[string]Pmapi

// Pocketmine-PMMP versions
func (poggit Poggit) PMapi() (Pmapis, error) {
	rootUrl := poggit.Host
	if rootUrl == nil {
		rootUrl = DefaultAPI
	}
	headers := request.Header{}
	data, _, err := request.JSON[Pmapis](rootUrl.ResolveReference(&url.URL{Path: path.Join(rootUrl.Path, "pmapis")}).String(), &request.Options{Header: headers})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Plugin struct
type Plugin struct {
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	Version        string              `json:"versions"`
	Html           string              `json:"html_url"`
	Tagline        string              `json:"tagline"`
	Artifact       string              `json:"artifact_url"`
	Downloads      int                 `json:"downloads"`
	Score          int                 `json:"score"`
	RepoID         int                 `json:"repo_id"`
	RepoName       string              `json:"repo_name"`
	ProjectID      int                 `json:"project_id"`
	ProjectName    string              `json:"project_name"`
	BuildID        int                 `json:"build_id"`
	Build          int                 `json:"build_number"`
	BuildCommit    string              `json:"build_commit"`
	DescriptionURL string              `json:"description_url"`
	IconURL        string              `json:"icon_url"`
	Changelog      string              `json:"changelog_url"`
	License        string              `json:"license"`
	LicenseURL     string              `json:"license_url"`
	Obsolete       bool                `json:"is_obsolete"`
	PreRelease     bool                `json:"is_pre_release"`
	Outdated       bool                `json:"is_outdated"`
	Official       bool                `json:"is_official"`
	Abandoned      bool                `json:"is_abandoned"`
	Submission     time.Time           `json:"submission_date"`
	LastChange     time.Time           `json:"last_state_change_date"`
	State          string              `json:"state_name"`
	Keywords       []string            `json:"keywords"`
	Producers      map[string][]string `json:"producers"`
	Categories     []struct {
		Major    bool   `json:"major"`
		Category string `json:"category_name"`
	} `json:"categories"`
	Dependencies []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		ID      int    `json:"depRelId"`
		IsHard  bool   `json:"isHard"`
	} `json:"deps"`
}

// List all plugins if possible
func (poggit Poggit) Plugins() ([]Plugin, error) {
	rootUrl := poggit.Host
	if rootUrl == nil {
		rootUrl = DefaultAPI
	}
	headers := request.Header{}

	data, _, err := request.JSON[[]Plugin](rootUrl.ResolveReference(&url.URL{Path: path.Join(rootUrl.Path, "plugins.min.json")}).String(), &request.Options{
		Header: headers,
		CodeProcess: request.MapCode{
			404: func(res *http.Response) (*http.Response, error) {
				if res != nil && res.Body != nil {
					res.Body.Close()
				}
				return request.MakeRequest(rootUrl.ResolveReference(&url.URL{Path: path.Join(rootUrl.Path, "plugins.json")}), &request.Options{Header: headers})
			},
		},
	})

	if err != nil {
		return nil, err
	}
	return data, nil
}

// Download plugin Phar file
func (poggit Poggit) DownloadPhar(resourceId int) (io.ReadCloser, error) {
	rootUrl := poggit.Host
	if rootUrl == nil {
		rootUrl = DefaultAPI
	}
	headers := request.Header{}
	res, err := request.MakeRequestWithStatus(rootUrl.ResolveReference(&url.URL{Path: path.Join(rootUrl.Path, "r", strconv.Itoa(resourceId))}), &request.Options{Header: headers})
	if err != nil {
		return nil, err
	} else if contentType := res.Header.Values("Content-Type"); len(contentType) > 0 && !slices.Contains(contentType, "application/octet-stream") {
		defer res.Body.Close()
		return nil, fmt.Errorf("cannot get valid phar file")
	}
	return res.Body, nil
}
