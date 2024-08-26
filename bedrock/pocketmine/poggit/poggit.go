package poggit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"sirherobrine23.com.br/go-bds/go-bds/request"
)

const (
	APIPlugins string = "%s/plugins.json"
)

type plugin struct {
	ID                  int         `json:"id"`
	Name                string      `json:"name"`
	Version             string      `json:"version"`
	HTMLURL             string      `json:"html_url"`
	Tagline             string      `json:"tagline"`
	ArtifactURL         string      `json:"artifact_url"`
	Downloads           int         `json:"downloads"`
	Score               interface{} `json:"score"`
	RepoID              int         `json:"repo_id"`
	RepoName            string      `json:"repo_name"`
	ProjectID           int         `json:"project_id"`
	ProjectName         string      `json:"project_name"`
	BuildID             int         `json:"build_id"`
	BuildNumber         int         `json:"build_number"`
	BuildCommit         string      `json:"build_commit"`
	DescriptionURL      string      `json:"description_url"`
	IconURL             string      `json:"icon_url"`
	ChangelogURL        string      `json:"changelog_url"`
	License             string      `json:"license"`
	LicenseURL          interface{} `json:"license_url"`
	IsObsolete          bool        `json:"is_obsolete"`
	IsPreRelease        bool        `json:"is_pre_release"`
	IsOutdated          bool        `json:"is_outdated"`
	IsOfficial          bool        `json:"is_official"`
	IsAbandoned         bool        `json:"is_abandoned"`
	SubmissionDate      int         `json:"submission_date"`
	State               int         `json:"state"`
	LastStateChangeDate int         `json:"last_state_change_date"`
	Categories          []struct {
		Major        bool   `json:"major"`
		CategoryName string `json:"category_name"`
	} `json:"categories"`
	Keywords []string `json:"keywords"`
	API      []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"api"`
	Deps      []interface{} `json:"deps"`
	Producers struct {
		Collaborator []string `json:"Collaborator"`
	} `json:"producers"`
	StateName string `json:"state_name"`
}

type PluginVersion struct {
	ArtifactURL string `json:"artifact_url"`
	BuildCommit string `json:"commit"`
}

type Plugin struct {
	ID           int      `json:"id"`
	IsObsolete   bool     `json:"is_obsolete"`
	IsPreRelease bool     `json:"is_pre_release"`
	IsOutdated   bool     `json:"is_outdated"`
	IsOfficial   bool     `json:"is_official"`
	IsAbandoned  bool     `json:"is_abandoned"`
	Keywords     []string `json:"keywords"`
	Categories   []struct {
		Major        bool   `json:"major"`
		CategoryName string `json:"category_name"`
	} `json:"categories"`
	Versions map[string]PluginVersion `json:"versions"`
}

type Poggit struct {
	serverURL string
	Plugins   map[string]Plugin
}

func NewPoggitClient(poggitServer string) (Poggit, error) {
	urlInfo, err := url.Parse(poggitServer)
	if err != nil {
		return Poggit{}, err
	} else if !(urlInfo.Scheme == "http" || urlInfo.Scheme == "https" || urlInfo.Scheme == "socket") {
		return Poggit{}, fmt.Errorf("set valid url schema, examples: http, https or socket")
	}
	return Poggit{serverURL: poggitServer, Plugins: map[string]Plugin{}}, nil
}

func (version *PluginVersion) Getfile() (io.ReadCloser, error) {
	res, err := request.Request(request.RequestOptions{HttpError: true, Url: version.ArtifactURL})
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func (poggit *Poggit) ListPlugins() error {
	res, err := request.Request(request.RequestOptions{HttpError: true, Url: fmt.Sprintf(APIPlugins, poggit.serverURL)})
	if err != nil {
		return err
	}

	var releases []plugin
	defer res.Body.Close()
	if err = json.NewDecoder(res.Body).Decode(&releases); err != nil {
		return err
	}

	for _, pls := range releases {
		var pl Plugin
		var ext bool
		if pl, ext = poggit.Plugins[pls.Name]; !ext {
			pl = Plugin{
				ID:           pls.ID,
				IsObsolete:   pls.IsObsolete,
				IsPreRelease: pls.IsPreRelease,
				IsOutdated:   pls.IsOutdated,
				IsOfficial:   pls.IsOfficial,
				IsAbandoned:  pls.IsAbandoned,
				Keywords:     pls.Keywords,
				Categories:   pls.Categories,
				Versions:     map[string]PluginVersion{},
			}
		}

		pl.IsObsolete = pls.IsObsolete
		pl.IsPreRelease = pls.IsPreRelease
		pl.IsOutdated = pls.IsOutdated
		pl.IsOfficial = pls.IsOfficial
		pl.IsAbandoned = pls.IsAbandoned
		pl.Keywords = pls.Keywords
		pl.Versions[pls.Version] = PluginVersion{
			ArtifactURL: pls.ArtifactURL,
			BuildCommit: pls.BuildCommit,
		}

		poggit.Plugins[pls.Name] = pl
	}

	return nil
}
