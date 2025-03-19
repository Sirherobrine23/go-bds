package bedrock

import (
	"bufio"
	"io"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/logs"
	"sirherobrine23.com.br/go-bds/go-bds/utils/slice"
)

var (
	_ = logs.RegisterParse[*BedrockParse]("mojang/bedrock")

	_ logs.Log    = &BedrockParse{}
	_ logs.Player = &BedrockPlayer{}
)

type BedrockPlayer struct {
	Username   string      `json:"player"` // Player username
	Actioned   logs.Action `json:"action"` // Action type
	Timed      time.Time   `json:"time"`   // Action time
	PlayerXUID int64       `json:"xuid"`   // Player xuid
	PFID       string      `json:"pfid"`
}

func (player BedrockPlayer) Name() string        { return player.Username }
func (player BedrockPlayer) Action() logs.Action { return player.Actioned }
func (player BedrockPlayer) Time() time.Time     { return player.Timed }
func (player BedrockPlayer) XUID() int64         { return player.PlayerXUID }

type BedrockParse struct {
	LastCompaction time.Time                `json:"last_world_compaction"` // Last world size reduce
	SessionID      string                   `json:"session"`               // Server session
	CommitID       string                   `json:"commit"`                // commit hash
	Branch         string                   `json:"branch"`                // Server Branch build
	ServerPlaform  *logs.Server             `json:"info"`                  // Basic server info
	Players        map[string][]logs.Player `json:"players"`               // Players
	Errs           []error                  `json:"errors"`
	Warngs         []error                  `json:"warnings"`

	err error
}

func (bedrock BedrockParse) Server() *logs.Server { return bedrock.ServerPlaform }
func (bedrock BedrockParse) Errors() []error      { return bedrock.Errs }
func (bedrock BedrockParse) Warnings() []error    { return bedrock.Warngs }
func (bedrock BedrockParse) GetPlayer(name string) (player []logs.Player, ok bool) {
	player, ok = bedrock.Players[name]
	return
}

func (bedrock *BedrockParse) ParseTime(current time.Time, log io.Reader) error {
	return bedrock.Parse(log)
}
func (bedrock *BedrockParse) Parse(log io.Reader) error {
	bedrock.ServerPlaform = &logs.Server{Platform: "mojang/bedrock", Ports: []*logs.Port{}} // Init info
	bedrock.Errs, bedrock.Warngs = []error{}, []error{}
	bedrock.Players = map[string][]logs.Player{}

	scanner := bufio.NewScanner(log)
	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimPrefix(text, "NO LOG FILE! - ")
		if strings.HasPrefix(text, "setting up server") || strings.HasPrefix(text, "Quit correctly") || text == "" {
			continue
		} else if !(strings.Contains(text, "]")) {
			return logs.ErrSkipPlatform
		}

		prefixEnd := strings.Index(text, "]")
		prefix := strings.TrimSpace(strings.TrimSuffix(text[strings.Index(text, "[")+1:prefixEnd], "INFO"))
		line := strings.TrimSpace(text[prefixEnd+1:])
		if line == "" {
			continue // skip
		} else if len(prefix) < 19 {
			return logs.ErrSkipPlatform
		}

		err := error(nil)
		if strings.HasSuffix(prefix, "ERROR") || strings.HasSuffix(prefix, "WARN") {
			bedrock.Errs = append(bedrock.Errs, &logs.ErrorReference{LogLevel: 1, FistLine: line})
			continue
		}

		// Date
		if strings.Count(prefix, ":") == 3 {
			lastColonIndex := strings.LastIndex(prefix, ":")
			prefix = prefix[:lastColonIndex] + "." + prefix[lastColonIndex+1:]
		}
		EntryTime, err := time.Parse("2006-01-02 15:04:05.999", prefix)
		if err != nil {
			if EntryTime, err = time.Parse("2006-01-02 15:04:05", prefix[:19]); err != nil {
				return err
			}
		}
		EntryTime = EntryTime.UTC() // Convert to UTC time

		explodeString := slice.Slice[string](strings.Fields(line))
		switch explodeString.At(0) {
		case "Server":
			if strings.Contains(text, "started") {
				bedrock.ServerPlaform.Started = EntryTime
			}
		case "Session":
			if explodeString.At(1) == "ID" {
				bedrock.SessionID = explodeString.At(2)
			}
		case "Branch:":
			bedrock.Branch = explodeString.At(1)
		case "Commit":
			bedrock.CommitID = explodeString.At(2)
		case "Running":
			if strings.HasPrefix(explodeString.At(1), "AutoCompaction") {
				bedrock.LastCompaction = EntryTime
			}
		case "Version", "Version:":
			if len(explodeString) >= 2 {
				bedrock.ServerPlaform.Version = explodeString.At(-1)
			}
		case "IPv6", "IPv4", "Listening":
			if explodeString.At(-2) == "port:" {
				protoLocation := 0
				if explodeString.At(0) == "Listening" {
					protoLocation = len(explodeString) - 3
				}

				port, err := strconv.ParseInt(explodeString.At(-1), 10, 16)
				if err != nil {
					return err
				}

				addr := netip.AddrPortFrom(netip.IPv4Unspecified(), uint16(port))
				if explodeString[protoLocation] == "IPv6" {
					addr = netip.AddrPortFrom(netip.IPv6Unspecified(), uint16(port))
				}

				bedrock.ServerPlaform.Ports = append(bedrock.ServerPlaform.Ports, &logs.Port{
					AddrPort: addr,
					From:     "server",
				})
			}
		case "Player":
			// Player connected:
			// Player disconnected:
			// Player connected: 2535413418839840
			// Player disconnected: 2535413418839840

			// Player connected: Sirherobrine, xuid: 2535413418839840
			// Player Spawned: Sirherobrine xuid: 2535413418839840
			// Player disconnected: Sirherobrine, xuid: 2535413418839840

			// Player Connected: nod dd, xuid: , pfid: c31902da495f4549
			// Player Spawned: nod dd xuid: , pfid: c31902da495f4549
			// Player disconnected: nod dd, xuid: , pfid: c31902da495f4549
			if !slices.Contains([]string{"connected:", "spawned:", "disconnected:"}, strings.ToLower(explodeString.At(1))) {
				continue
			}

			var player, xuid, pfid string
			if player = strings.TrimSpace(line[strings.Index(line, explodeString.At(1))+len(explodeString.At(1)):]); player == "" {
				// Old versions of Minecraft Bedrock Server (before 1.6.0)
				// did not return the names of those who were not logged in
				continue
			}

			if strings.Contains(player, ", xuid:") {
				xuidd := strings.SplitN(player, ", xuid:", 2)
				player = strings.TrimSpace(xuidd[0])
				xuid = strings.TrimSpace(xuidd[1])
				if strings.Contains(xuid, ", pfid:") {
					pdis := strings.SplitN(xuid, ", pfid:", 2)
					xuid = strings.TrimSpace(pdis[0])
					pfid = strings.TrimSpace(pdis[1])
				}
			}

			if strings.Contains(player, "xuid:") {
				xuid = player[strings.LastIndex(player, "xuid:")+5:]
				player = player[:strings.LastIndex(player, "xuid:")-1]
				if player[len(player)-1] == ',' {
					player = player[:len(player)-1]
				}
				if strings.Contains(xuid, ",") {
					xuid = xuid[:strings.Index(player, ",")]
				}
				xuid = strings.TrimSpace(xuid)
				player = strings.TrimSpace(player)
			}

			level := logs.Action(0)
			switch strings.ToLower(explodeString.At(1)) {
			case "disconnected:":
				level = logs.Disconnect
			case "connected:":
				level = logs.Connect
			case "spawned:":
				level = logs.Spawned
			}

			if _, ok := bedrock.Players[player]; !ok {
				bedrock.Players[player] = []logs.Player{}
			}

			xuidID, _ := strconv.ParseInt(xuid, 10, 64) // Convert xuid string to XUID int
			bedrock.Players[player] = append(bedrock.Players[player], BedrockPlayer{
				Username:   player,
				Actioned:   level,
				Timed:      EntryTime,
				PlayerXUID: xuidID,
				PFID:       pfid,
			})
		}
	}

	return scanner.Err()
}
