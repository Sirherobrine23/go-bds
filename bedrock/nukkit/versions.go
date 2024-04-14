package nukkit

import (
	"encoding/json"
	"net/url"
	"time"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	MasterNukkit string = "https://ci.opencollab.dev/job/NukkitX/job/Nukkit/job/master/api/json"
)

type NukkitBuild struct {
	CommitID    string    `json:"commit"`
	BuilderDate time.Time `json:"buildDate"`
	FileUrl     string    `json:"url"`
}

func ListFiles() ([]NukkitBuild, error) {
	res, err := request.Request(request.RequestOptions{HttpError: true, Url: MasterNukkit})
	if err != nil {
		return []NukkitBuild{}, err
	}
	defer res.Body.Close()

	var builds struct {
		Builds []struct {
			BuilderNumber int64  `json:"number"`
			BuilderPage   string `json:"url"`
		} `json:"builds"`
	}
	if err = json.NewDecoder(res.Body).Decode(&builds); err != nil {
		return []NukkitBuild{}, err
	}

	if len(builds.Builds) == 0 {
		return []NukkitBuild{}, nil
	}
	releasesData := []NukkitBuild{}

	page, err := url.Parse(builds.Builds[0].BuilderPage)
	if err != nil {
		return []NukkitBuild{}, err
	}
	page.Path, _ = url.JoinPath(page.Path, "api/json")

	pageApi := page.String()
	for {
		if len(pageApi) == 0 {
			break
		}

		res, err := request.Request(request.RequestOptions{HttpError: true, Url: pageApi})
		if err != nil {
			return []NukkitBuild{}, err
		}
		defer res.Body.Close()

		var result struct {
			Timestamp int64  `json:"timestamp"`
			Result    string `json:"result"`
			PageUrl   string `json:"url"`
			Previous  struct {
				Page string `json:"url"`
			} `json:"previousBuild"`
			Artifacts []struct {
				Display  string `json:"displayPath"`
				Filename string `json:"fileName"`
				Path     string `json:"relativePath"`
			} `json:"artifacts"`
			Actions []struct {
				Class   string `json:"_class"`
				Builder map[string]struct {
					Marked struct {
						SHA1 string `json:"revision"`
					} `json:"marked"`
					Revision struct {
						SHA1 string `json:"revision"`
					} `json:"revision"`
				} `json:"buildsByBranchName"`
			} `json:"actions"`
		}
		if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
			return releasesData, err
		}

		pageApi = ""
		if len(result.Previous.Page) > 5 {
			page, err := url.Parse(result.Previous.Page)
			if err != nil {
				return []NukkitBuild{}, err
			}
			page.Path, _ = url.JoinPath(page.Path, "api/json")
			pageApi = page.String()
		}

		for _, at := range result.Artifacts {
			filePath, err := url.Parse(result.PageUrl)
			if err != nil {
				return releasesData, err
			}

			filePath.Path, _ = url.JoinPath(filePath.Path, at.Path)
			var sha1 string
			for _, k := range result.Actions {
				if k.Class == "hudson.plugins.git.util.BuildData" {
					sha1 = k.Builder["master"].Marked.SHA1
					break
				}
			}
			releasesData = append(releasesData, NukkitBuild{
				CommitID:    sha1,
				BuilderDate: time.UnixMilli(result.Timestamp),
				FileUrl:     filePath.String(),
			})
		}
	}

	return releasesData, nil
}
