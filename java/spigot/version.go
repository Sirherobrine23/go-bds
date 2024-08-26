package spigot

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"code.gitea.io/sdk/gitea"
	"sirherobrine23.com.br/go-bds/go-bds/request"
)

type Version struct {
	Version     string    // Version
	ReleaseDate time.Time // Build date
	Downloads   []string  // Files to Download
}

func (ver *Version) Download(InstallPath string) error {
	for _, fileUrl := range ver.Downloads {
		savePath := filepath.Join(InstallPath, filepath.Base(fileUrl))
		res, err := (&request.RequestOptions{Url: fileUrl}).Request()
		if err != nil {
			return err
		}
		defer res.Body.Close()
		file, err := os.Open(savePath)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := io.Copy(file, res.Body); err != nil {
			return err
		}
	}
	return nil
}

func GetReleases() ([]Version, error) {
	tea, err := gitea.NewClient("https://sirherobrine23.com.br")
	if err != nil {
		return nil, err
	}
	versions := []Version{}
	page := 0
	for {
		res, teaRes, err := tea.ListReleases("go-bds", "Spigot", gitea.ListReleasesOptions{
			ListOptions: gitea.ListOptions{
				PageSize: 100000,
				Page:     page,
			},
		})
		page = teaRes.NextPage
		if err != nil {
			return nil, err
		}
		for _, release := range res {
			rel := Version{
				Version:     release.TagName,
				ReleaseDate: release.CreatedAt,
				Downloads:   []string{},
			}
			for _, attach := range release.Attachments {
				rel.Downloads = append(rel.Downloads, attach.DownloadURL)
			}
			versions = append(versions, rel)
		}
		if page == teaRes.LastPage {
			break
		}
	}
	return versions, nil
}
