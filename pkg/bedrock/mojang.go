package bedrock

import (
	"time"

	"sirherobrine23.org/minecraft-server/go-bds/internal/request"
)

type BedrockVersionsTarget struct {
	NodePlatform string `json:"Platform"`
	NodeArch     string `json:"Arch"`
	ZipFile      string `json:"zip"`
	TarFile      string `json:"tar"`
	ZipSHA1      string `json:"zipSHA1"`
	TarSHA1      string `json:"tarSHA1"`
}

type BedrockVersions struct {
	Version     string                  `json:"version"`
	DateRelease time.Time               `json:"releaseDate"`
	ReleaseType string                  `json:"type"`
	Targets     []BedrockVersionsTarget `json:"targets"`
}

// List Minecraft bedrock server versions and file Urls
func GetMojangVersions() ([]BedrockVersions, error) {
	var data []BedrockVersions
	err := request.GetJson("https://sirherobrine23.org/Minecraft-Server/BedrockFetch/raw/branch/main/versions.json", &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func RunBedrock(serverPath string) {}