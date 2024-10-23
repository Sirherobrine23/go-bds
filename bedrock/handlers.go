package bedrock

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/internal/regex"
)

const (
	PlayerActionDisconnect string = "disconnect" // Player disconnected from server
	PlayerActionConnect    string = "connect"    // Player connect in to server
	PlayerActionSpawn      string = "spawn"      // Player spawned in server and connected correct to server, only new server (1.16+)
)

var (
	MojangPlayerActions = map[string]*regex.Regexp{
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
		`v2`: regex.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\ Player (?P<Action>disconnected|connected|Spawned): (?P<Username>[0-9A-Za-z_\-\s]+), xuid:\s?(?P<Xuid>[0-9A-Za-z]+)?,?`),
	}
	MojangPort = map[string]*regex.Regexp{
		// NO LOG FILE! - [2024-07-30 00:33:21 INFO] IPv4 supported, port: 19132
		// NO LOG FILE! - [2024-07-30 00:33:21 INFO] IPv6 supported, port: 19133
		// NO LOG FILE! - [2024-07-30 00:33:22 INFO] Listening on IPv6 port: 19133
		// NO LOG FILE! - [2024-07-30 00:33:22 INFO] Listening on IPv4 port: 19132
		//                [2023-03-08 13:01:57 INFO] Listening on IPv4 port: 19132
		`v3`: regex.MustCompile(`(?m)\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})) INFO\] Listening on IPv(?P<Protocol>4|6) port: (?P<Port>[0-9]{1,5})`),
		// [2024-07-29 20:48:07:066 INFO] IPv4 supported, port: 19132: Used for gameplay and LAN discovery
		// [2024-07-29 20:48:07:066 INFO] IPv6 supported, port: 19133: Used for gameplay
		`v2`: regex.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\] IPv(?P<Protocol>4|6) supported, port: (?P<Port>[0-9]{1,5})`),
		// [INFO] IPv4 supported, port: 19132
		`v1`: regex.MustCompile(`(?m)^\[INFO\] IPv(?P<Protocol>4|6) supported, port: (?P<Port>[0-9]{1,5})`),
	}
	MojangStarter = map[string]*regex.Regexp{
		// [2024-04-10 11:16:29:640 INFO] Server started.
		`v2`: regex.MustCompile(`(?m)^\[(?P<TimeAction>([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})):[0-9]{1,3} INFO\] Server started\.`),
	}
)

type PlayerConnection struct {
	Player         string    `json:"player"`         // Player username
	XUID           string    `json:"xuid,omitempty"` // Player xuid
	Action         string    `json:"action"`         // Connection type
	TimeConnection time.Time `json:"connectionTime"` // Player connection time
}

type Handlers struct {
	Started *time.Time         `json:"started"` // Server started date
	Ports   []uint16           `json:"ports"`   // Server ports
	Players []PlayerConnection `json:"players"` // Player connnections in to server
}

// Server avaible time to player connect
func (w *Handlers) ParseStarted(logline string) {
	for _, reg := range MojangStarter {
		if reg.MatchString(logline) {
			var err error
			matched := reg.FindAllGroups(logline)
			w.Started = new(time.Time)
			if timeStr, ok := matched["TimeAction"]; ok {
				if *w.Started, err = time.ParseInLocation(`2006-01-02 15:04:05`, timeStr, time.Local); err == nil {
					return
				}
			} else if timeStr, ok := matched["Time"]; ok {
				if *w.Started, err = time.ParseInLocation(`2006-01-02 15:04:05`, timeStr, time.Local); err == nil {
					return
				}
			}
			*w.Started = time.Now()
			return
		}
	}
}

// Player action
func (w *Handlers) ParsePlayer(logline string) {
	for _, reg := range MojangPlayerActions {
		if reg.MatchString(logline) {
			ActionPlayer := reg.FindAllGroups(logline)
			var timed = time.Now()
			if timeAct, ok := ActionPlayer["TimeAction"]; ok {
				var err error
				if timed, err = time.Parse(`2006-01-02 15:04:05`, timeAct); err != nil {
					return
				}
			}

			action := strings.ToLower(ActionPlayer["Action"])
			if slices.Contains([]string{PlayerActionConnect, PlayerActionDisconnect, PlayerActionSpawn}, action) {
				w.Players = append(w.Players, PlayerConnection{
					Player:         strings.TrimSpace(ActionPlayer["Username"]),
					XUID:           strings.TrimSpace(ActionPlayer["Xuid"]),
					Action:         action,
					TimeConnection: timed,
				})
			}
			return
		}
	}
}

// Port listen
func (w *Handlers) ParsePort(logline string) {
	for _, reg := range MojangPort {
		if reg.MatchString(logline) {
			infoPort := reg.FindAllGroups(logline)
			var port uint16
			if _, err := fmt.Sscan(infoPort["Port"], &port); err != nil {
				return
			}

			// Check if port ared added to slice
			if !slices.Contains(w.Ports, port) {
				w.Ports = append(w.Ports, port)
			}
			return
		}
	}
}

// Parse log and register on handlers
func (w *Handlers) RegisterScan(log io.ReadCloser) {
	defer log.Close()
	logScan := bufio.NewScanner(log)
	for logScan.Scan() {
		logLine := logScan.Text()
		go w.ParseStarted(logLine)
		go w.ParsePlayer(logLine)
		go w.ParsePort(logLine)
	}
}
