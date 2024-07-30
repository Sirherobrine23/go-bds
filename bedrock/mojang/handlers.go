package mojang

import (
	"bufio"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"sirherobrine23.org/go-bds/go-bds/internal"
)

const (
	PlayerActionDisconnect string = "disconnect" // Player disconnected from server
	PlayerActionConnect    string = "connect"    // Player connect in to server
	PlayerActionSpawn      string = "spawn"      // Player spawned in server and connected correct to server, only new server (1.16+)
)

/* v1.6.1.0:
NO LOG FILE! - setting up server logging...
NO LOG FILE! - [2024-07-30 00:33:20 INFO] Starting Server
NO LOG FILE! - [2024-07-30 00:33:20 INFO] Version 1.6.1.0
NO LOG FILE! - [2024-07-30 00:33:20 INFO] Level Name: Bedrock level
NO LOG FILE! - [2024-07-30 00:33:21 ERROR] Error opening whitelist file: whitelist.json
NO LOG FILE! - [2024-07-30 00:33:21 ERROR] Error opening ops file: ops.json
NO LOG FILE! - [2024-07-30 00:33:21 INFO] Game mode: 0 Survival
NO LOG FILE! - [2024-07-30 00:33:21 INFO] Difficulty: 1 EASY
NO LOG FILE! - [2024-07-30 00:33:21 INFO] IPv4 supported, port: 19132
NO LOG FILE! - [2024-07-30 00:33:21 INFO] IPv6 supported, port: 19133
NO LOG FILE! - [2024-07-30 00:33:22 INFO] Listening on IPv6 port: 19133
NO LOG FILE! - [2024-07-30 00:33:22 INFO] Listening on IPv4 port: 19132
*/

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
		`v2`: regexp.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\ Player (?P<Action>disconnected|connected|Spawned): (?P<Username>[0-9A-Za-z_\-\s]+), xuid:\s?(?P<Xuid>[0-9A-Za-z]+)?,?`),
	}
	MojangPort = map[string]*regexp.Regexp{
		// [2023-03-08 13:01:57 INFO] Listening on IPv4 port: 19132
		`v3`: regexp.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})) INFO\] Listening on IPv(?P<Protocol>4|6) port: (?P<Port>[0-9]{1,5})`),
		// [2024-07-29 20:48:07:066 INFO] IPv4 supported, port: 19132: Used for gameplay and LAN discovery
		// [2024-07-29 20:48:07:066 INFO] IPv6 supported, port: 19133: Used for gameplay
		`v2`: regexp.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\] IPv(?P<Protocol>4|6) supported, port: (?P<Port>[0-9]{1,5})`),
		// [INFO] IPv4 supported, port: 19132
		`v1`: regexp.MustCompile(`(?m)^\[INFO\] IPv(?P<Protocol>4|6) supported, port: (?P<Port>[0-9]{1,5})`),
	}
	MojangStarter = map[string]*regexp.Regexp{
		// [2024-04-10 11:16:29:640 INFO] Server started.
		`v2`: regexp.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\] Server started\.`),
	}
)

type PlayerConnection struct {
	XUID           string    `json:"xuid,omitempty"` // Player xuid
	Action         string    `json:"action"`         // Connection type
	TimeConnection time.Time `json:"connectionTime"` // Player connection time
}

type Handlers struct {
	Started         time.Time                                            // Server started date
	Ports           []uint16                                             // Server ports
	Players         map[string][]PlayerConnection                        // Player connnections in to server
	playerCallbacks []func(Username string, PlayerInfo PlayerConnection) // Callbacks to player
}

func (w *Handlers) PlayerAction(callback func(Username string, PlayerInfo PlayerConnection)) *Handlers {
	w.playerCallbacks = append(w.playerCallbacks, callback)
	return w
}

// Parse log and register on handlers
func (w *Handlers) RegisterScan(log io.ReadCloser) {
	defer log.Close()
	logScan := bufio.NewScanner(log)
	for logScan.Scan() {
		logLine := logScan.Text()

		// Started time
		if MojangStarter["v2"].MatchString(logLine) {
			TimedString := internal.FindAllGroups(MojangStarter["v2"], logLine)["TimeAction"]
			var err error
			w.Started, err = time.ParseInLocation(`2006-01-02 15:04:05`, TimedString, time.Local)
			if err != nil {
				w.Started = time.Now()
			}
			continue
		}

		// Port listen
		if MojangPort["v1"].MatchString(logLine) {
			infoPort := internal.FindAllGroups(MojangPort["v1"], logLine)
			port, err := strconv.Atoi(infoPort["Port"])
			if err != nil {
				continue
			}
			w.Ports = append(w.Ports, uint16(port))
			continue
		} else if MojangPort["v2"].MatchString(logLine) {
			infoPort := internal.FindAllGroups(MojangPort["v2"], logLine)
			port, err := strconv.Atoi(infoPort["Port"])
			if err != nil {
				continue
			}
			w.Ports = append(w.Ports, uint16(port))
			continue
		}

		// Player action
		if MojangPlayerActions["v2"].MatchString(logLine) {
			ActionPlayer := internal.FindAllGroups(MojangPlayerActions["v2"], logLine)
			var err error
			var Player PlayerConnection
			if Player.TimeConnection, err = time.Parse(`2006-01-02 15:04:05`, ActionPlayer["TimeAction"]); err != nil {
				continue
			}
			Player.XUID = strings.TrimSpace(ActionPlayer["Xuid"])
			Player.Action = strings.ToLower(ActionPlayer["Action"])
			Username := ActionPlayer["Username"]

			// Callback
			if slices.Contains([]string{PlayerActionConnect, PlayerActionDisconnect, PlayerActionSpawn}, Player.Action) {
				if _, ext := w.Players[Username]; !ext {
					w.Players[Username] = []PlayerConnection{}
				}

				w.Players[Username] = append(w.Players[Username], Player)
				for _, callback := range w.playerCallbacks {
					go callback(Username, Player)
				}
			}
			continue
		}
	}
}
