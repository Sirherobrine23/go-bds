package bedrock

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"sirherobrine23.org/minecraft-server/go-bds/internal/request"
)

var MojangPlayerActions = map[string]*regexp.Regexp{
	`v1`: nil,
	// [2024-04-01 20:50:26:198 INFO] Player connected: Sirherobrine, xuid: 2535413418839840
	// [2024-04-01 21:46:11:691 INFO] Player connected: nod dd, xuid:
	// [2024-04-01 20:50:31:386 INFO] Player Spawned: Sirherobrine xuid: 2535413418839840
	// [2024-04-01 21:46:16:637 INFO] Player Spawned: nod dd xuid: , pfid: c31902da495f4549
	// [2022-08-30 20:56:55:231 INFO] Player disconnected: Sirherobrine, xuid: 2535413418839840
	// [2024-04-01 21:46:33:199 INFO] Player disconnected: nod dd, xuid: , pfid: c31902da495f4549
	`v2`: regexp.MustCompile(`(?m)^\[.*\] Player (?P<Action>disconnected|connected|Spawned): (?P<Username>[0-9A-Za-z_\-\s]+), xuid:\s?(?P<xuid>[0-9A-Za-z]+)?,?`),
}

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
	return data, err
}

type Player struct {
	Username string
	Action   string
	Xuid     string
}

func ParseBedrockPlayerAction(line string) (Player, error) {
	if MojangPlayerActions["v2"].MatchString(line) {
		match := MojangPlayerActions["v2"].FindStringSubmatch(line)
		if len(match) >= 4 && strings.TrimSpace(match[3]) != "" {
			return Player{
				match[2],
				strings.ToLower(match[1]),
				strings.TrimSpace(match[3]),
			}, nil
		}
		return Player{
			match[2],
			strings.ToLower(match[1]),
			"",
		}, nil
	}

	return Player{}, fmt.Errorf("cannot get player info")
}