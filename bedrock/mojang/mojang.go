package mojang

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/properties"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/exec"
)

const (
	PlayerActionDisconnect string = "disconnect" //
	PlayerActionConnect    string = "connect"    //
	PlayerActionSpawn      string = "spawn"      //
)

var (
	MojangPlayerActions = map[string]*regexp.Regexp{
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
	MojangPort = map[string]*regexp.Regexp{
		// [2023-03-08 13:01:57 INFO] Listening on IPv4 port: 19132
		`v2`: regexp.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\] Listening on IPv(?P<Protocol>4|6) port: (?P<Port>[0-9]+)$`),
		// [INFO] IPv4 supported, port: 19132
		`v1`: regexp.MustCompile(`(?m)^\[INFO\] IPv(?P<Protocol>4|6) supported, port: (?P<Port>[0-9]+)$`),
	}
	MojangStarter = map[string]*regexp.Regexp{
		// [2024-04-10 11:16:29:640 INFO] Server started.
		`v2`: regexp.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\] Server started\.`),
	}
)

type PlayerConnections struct {
	XUID           string    `json:"xuid,omitempty"` // Player xuid
	Action         string    `json:"action"`         // Connection type
	TimeConnection time.Time `json:"connectionTime"` // Player connection time
}

type Mojang struct {
	ServerPath      string                                                `json:"serverPath"`    // Server path to download, run server
	Version         string                                                `json:"serverVersion"` // Server version
	Started         time.Time                                             `json:"startedTime"`   // Server started date
	Ports           []int                                                 `json:"ports"`         // Server ports
	Players         map[string][]PlayerConnections                        `json:"players"`       // Player connnections in to server
	Config          MojangConfig                                          `json:"serverConfig"`  // Config server file
	playerCallbacks []func(Username string, PlayerInfo PlayerConnections) `json:"-"`             // Callbacks to player
}

func (w *Mojang) Download() (VersionTarget, error) {
	versions, err := FromVersions()
	if err != nil {
		return VersionTarget{}, err
	}

	goTarget := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	for _, ver := range versions {
		if ver.Version == w.Version {
			for _, target := range ver.Targets {
				if target.Target == goTarget {
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

func (w *Mojang) Start() (exec.Server, error) {
	data, err := os.ReadFile(filepath.Join(w.ServerPath, "server.properties"))
	if err != nil {
		return exec.Server{}, err
	} else if err = properties.Unmarshal(data, &w.Config); err != nil {
		return exec.Server{}, err
	}

	filename := "./bedrock_server"
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	exeProcess, err := exec.Run(exec.ServerOptions{
		Cwd:       w.ServerPath,
		Arguments: []string{filename},
	})

	if err != nil {
		return exeProcess, err
	}

	var log io.ReadCloser
	if log, err = exeProcess.Stdlog.NewPipe(); err != nil {
		return exeProcess, err
	}

	lineBreaker := bufio.NewScanner(log)
	go (func() {
		for lineBreaker.Scan() {
			line := lineBreaker.Text()

			// Started time
			go (func() {
				if MojangStarter["v2"].MatchString(line) {
					TimedString := internal.FindAllGroups(MojangStarter["v2"], line)["TimeAction"]
					w.Started, _ = time.Parse(`2006-01-02 15:04:05`, TimedString)
				}
			})()

			// Port listen
			go (func() {
				var infoPort map[string]string
				if MojangPort["v1"].MatchString(line) {
					infoPort = internal.FindAllGroups(MojangPort["v1"], line)
				} else if MojangPort["v2"].MatchString(line) {
					infoPort = internal.FindAllGroups(MojangPort["v2"], line)
				} else {
					return
				}

				if infoPort["Protocol"] == "4" || infoPort["Protocol"] == "6" {
					port, err := strconv.Atoi(infoPort["Port"])
					if err != nil {
						return
					}
					w.Ports = append(w.Ports, port)
				}
			})()

			// Player action
			go (func() {
				if MojangPlayerActions["v2"].MatchString(line) {
					ActionPlayer := internal.FindAllGroups(MojangPlayerActions["v2"], line)

					Username := ActionPlayer["Username"]
					Action := strings.ToLower(ActionPlayer["Action"])
					Xuid := strings.TrimSpace(ActionPlayer["Xuid"])
					timed, err := time.Parse(`2006-01-02 15:04:05`, ActionPlayer["TimeAction"])

					if err != nil {
						return
					}

					// Callback
					if Action == PlayerActionConnect || Action == PlayerActionDisconnect || Action == PlayerActionSpawn {
						if _, ext := w.Players[Username]; !ext {
							w.Players[Username] = []PlayerConnections{}
						}

						act := PlayerConnections{Xuid, Action, timed}
						w.Players[Username] = append(w.Players[Username], act)
						for _, callback := range w.playerCallbacks {
							go callback(Username, act)
						}
					}
				}
			})()
		}
	})()

	return exeProcess, err
}

func (w *Mojang) PlayerAction(callback func(Username string, PlayerInfo PlayerConnections)) *Mojang {
	w.playerCallbacks = append(w.playerCallbacks, callback)
	return w
}
