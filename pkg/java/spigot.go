package java

import (
	"fmt"

	"sirherobrine23.org/minecraft-server/go-bds/internal/request"
)

type SpigotVersion struct {
	Version     string `json:"version"`
	ServerUrl   string `json:"spigotUrl"`
	Craftbukkit string `json:"craftbukkitUrl"`
}

func GetSpigotVersions() ([]SpigotVersion, error) {
	versions := make([]SpigotVersion, 0)
	page := 0
	for {
		var data []struct {
			TagName string `json:"tag_name"`
			Files   []struct {
				Name    string `json:"name"`
				FileUrl string `json:"browser_download_url"`
			} `json:"assets"`
		}

		err := request.GetJson(fmt.Sprintf("https://sirherobrine23.org/api/v1/repos/Minecraft-Server/Spigot/releases?page=%d", page), &data)
		if err != nil {
			return nil, err
		} else if len(data) == 0 {
			break
		}
		page++
		for _, v := range data {
			if len(v.Files) >= 2 {
				file1 := v.Files[0]
				file2 := v.Files[1]
				if len(file2.Name) > 0 {
					if file1.Name == "server.jar" {
						versions = append(versions, SpigotVersion{
							Version: v.TagName,
							ServerUrl: file1.FileUrl,
							Craftbukkit: file2.FileUrl,
						})
					} else {
						versions = append(versions, SpigotVersion{
							Version: v.TagName,
							ServerUrl: file2.FileUrl,
							Craftbukkit: file1.FileUrl,
						})
					}
				}
			} else {
				versions = append(versions, SpigotVersion{
					Version: v.TagName,
					ServerUrl: v.Files[0].FileUrl,
				})
			}
		}
	}

	// return versions, nil
	return versions, nil
}
