package mojang

import (
	"path/filepath"
	"time"

	"code.gitea.io/sdk/gitea"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

type spigotAsset struct {
	Version     string
	ReleaseDate time.Time
	Files       []*gitea.Attachment
}

// Get all releases from sirherobrine23.com.br
//
// se https://www.spigotmc.org/wiki/faq/#where-do-i-get-the-spigot-or-bukkit-jar-file-for-my-server
func SpigotReleases() (Releases, error) {
	tea, err := gitea.NewClient("https://sirherobrine23.com.br")
	if err != nil {
		return nil, err
	}
	page := 0
	var assests []*gitea.Release
	for {
		var pageAssets []*gitea.Release
		var res *gitea.Response
		if pageAssets, res, err = tea.ListReleases("go-bds", "Spigot", gitea.ListReleasesOptions{ListOptions: gitea.ListOptions{PageSize: 100000, Page: page}}); err != nil {
			return nil, err
		}
		page = res.NextPage
		assests = append(assests, pageAssets...)
		if page == res.LastPage {
			break
		}
	}

	releases := make(Releases)
	for _, asset := range assests {
		releases[asset.TagName] = spigotAsset{
			Version:     asset.TagName,
			ReleaseDate: asset.CreatedAt,
			Files:       asset.Attachments,
		}
	}
	return releases, nil
}

func (rel spigotAsset) ReleaseType() string    { return "oficial" }
func (rel spigotAsset) String() string         { return rel.Version }
func (rel spigotAsset) ReleaseTime() time.Time { return rel.ReleaseDate }
func (rel spigotAsset) Download(folder string) error {
	for _, asset := range rel.Files {
		if _, err := request.SaveAs(asset.DownloadURL, filepath.Join(folder, asset.Name), nil); err != nil {
			return err
		}
	}
	return nil
}
