package mojang

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gookit/properties"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal"
)

const (
	PlayerActionDisconnect string = "disconnect" //
	PlayerActionConnect    string = "connect"    //
	PlayerActionSpawn      string = "spawn"      //
)

var MojangPlayerActions = map[string]*regexp.Regexp{
	// [2024-04-01 20:50:26:198 INFO] Player connected: Sirherobrine, xuid: 2535413418839840
	// [2024-04-01 21:46:11:691 INFO] Player connected: nod dd, xuid:
	// [2024-04-01 20:50:31:386 INFO] Player Spawned: Sirherobrine xuid: 2535413418839840
	// [2024-04-01 21:46:16:637 INFO] Player Spawned: nod dd xuid: , pfid: c31902da495f4549
	// [2022-08-30 20:56:55:231 INFO] Player disconnected: Sirherobrine, xuid: 2535413418839840
	// [2024-04-01 21:46:33:199 INFO] Player disconnected: nod dd, xuid: , pfid: c31902da495f4549
	//
	// TimeAction = time.Time{}
	// Action = disconnected|connected|Spawned
	// Username = String
	// Xuid = String
	`v2`: regexp.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\] Player (?P<Action>disconnected|connected|Spawned): (?P<Username>[0-9A-Za-z_\-\s]+), xuid:\s?(?P<Xuid>[0-9A-Za-z]+)?,?`),
	`v1`: nil,
}

type PlayerConnections struct {
	XUID           string    `json:"xuid,omitempty"` // Player xuid
	Action         string    `json:"action"`         // Connection type
	TimeConnection time.Time `json:"connectionTime"` // Player connection time
}

type Mojang struct {
	ServerPath string                         `json:"serverPath"`    // Server path to download, run server
	Version    string                         `json:"serverVersion"` // Server version
	Players    map[string][]PlayerConnections `json:"players"`       // Player connnections in to server
	Config     MojangConfig                   `json:"serverConfig"`  // Config server file
}

func ParseBedrockPlayerAction(line string, callback func(username string, playerInfo PlayerConnections)) error {
	if MojangPlayerActions["v2"].MatchString(line) {
		ActionPlayer := internal.FindAllGroups(MojangPlayerActions["v2"], line)

		Username := ActionPlayer["Username"]
		Action := ActionPlayer["Action"]
		Xuid := strings.TrimSpace(ActionPlayer["Xuid"])
		timed, err := time.Parse(`2006-01-02 15:04:05`, ActionPlayer["TimeAction"])

		if err != nil {
			return err
		}

		// Callback
		if PlayerActionConnect == Action || PlayerActionDisconnect == Action || PlayerActionSpawn == Action {
			callback(Username, PlayerConnections{Xuid, Action, timed})
			return nil
		}
	}

	return nil
}

func (w *Mojang) Download() (VersionTarget, error) {
	versions, err := FromVersions()
	if err != nil {
		return VersionTarget{}, err
	}

	for _, ver := range versions {
		if ver.Version == w.Version {
			for _, target := range ver.Targets {
				if runtime.GOOS == target.GOOS && runtime.GOARCH == target.GOARCH {
					return target, target.Download(w.ServerPath)
				}
			}
		}
	}

	return VersionTarget{}, ErrNoVersion
}

// Save config in to server.properties
func (w *Mojang) SaveConfig() error {
	data, err := properties.Marshal(&w.Config)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.ServerPath, "server.properties"), data, os.FileMode(0o666))
}

func (w *Mojang) Start() error {
	data, err := os.ReadFile(filepath.Join(w.ServerPath, "server.properties"))
	if err != nil {
		return err
	} else if err = properties.Unmarshal(data, &w.Config); err != nil {
		return err
	}

	return nil
}
