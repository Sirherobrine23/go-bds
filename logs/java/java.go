package java

import (
	"bufio"
	"io"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/logs"
	"sirherobrine23.com.br/go-bds/go-bds/regex"
	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
)

var (
	_             = logs.RegisterParse[*JavaParse]("mojang/java") // Register platform
	_ logs.Log    = (*JavaParse)(nil)
	_ logs.Player = (*JavaPlayer)(nil)

	DoneMatch *regex.Regexp = regex.MustCompile(`Done \([0-9\.]+s\)! For help, type "help"( or "\?")?`)
)

type JavaPlayer struct {
	Username string      `json:"player"` // Player username
	Actioned logs.Action `json:"action"` // Action type
	Timed    time.Time   `json:"time"`   // Action time
}

func (player JavaPlayer) Name() string        { return player.Username }
func (player JavaPlayer) Action() logs.Action { return player.Actioned }
func (player JavaPlayer) Time() time.Time     { return player.Timed }
func (player JavaPlayer) XUID() int64         { return -1 }

type JavaParse struct {
	ServerPlaform *logs.Server             `json:"info"`
	Players       map[string][]logs.Player `json:"players"`
	Errs          []error                  `json:"errors"`
	Warngs        []error                  `json:"warnings"`
}

func (java JavaParse) Server() *logs.Server { return java.ServerPlaform }
func (java JavaParse) Errors() []error      { return java.Errs }
func (java JavaParse) Warnings() []error    { return java.Warngs }
func (java JavaParse) GetPlayer(name string) (player []logs.Player, ok bool) {
	player, ok = java.Players[name]
	return
}

func (java *JavaParse) Parse(log io.Reader) error { return java.ParseTime(time.Now(), log) }
func (java *JavaParse) ParseTime(currentTime time.Time, log io.Reader) error {
	java.ServerPlaform = &logs.Server{Platform: "mojang/java", Ports: []*logs.Port{}} // Init info
	java.Errs, java.Warngs = []error{}, []error{}
	java.Players = map[string][]logs.Player{}

	valid, errorReference, scanner := false, error(nil), bufio.NewScanner(log)
	for scanner.Scan() {
		line := scanner.Text()
		valid = true
		if !(line[0] == 'U' || line[0] == 'S' || line[0] == '[') && errorReference != nil {
			errorReference.(*logs.ErrorReference).Line = append(errorReference.(*logs.ErrorReference).Line, line)
			continue
		} else if !(line[0] == 'U' || line[0] == 'S' || line[0] == '[') {
			valid = false
			break
		}

		if line[0] == 'U' || line[0] == 'S' {
			continue // Ignore line
		}

		prefixSplited := [3]string(strings.SplitAfterN(line, "]", 3))
		prefixSplited[0] = strings.Replace(strings.Replace(strings.TrimSpace(prefixSplited[0][1:]), "[", "", 1), "]", "", 1)
		prefixSplited[1] = strings.Replace(strings.Replace(strings.TrimSpace(prefixSplited[1][1:]), "[", "", 1), "]", "", 1)
		prefixSplited[2] = strings.TrimSpace(prefixSplited[2][1:])

		// Error log
		if strings.HasSuffix(prefixSplited[1], "WARN") || strings.HasSuffix(prefixSplited[1], "ERROR") {
			errorReference = &logs.ErrorReference{
				LogLevel: 1,
				FistLine: prefixSplited[2],
			}
			java.Warngs = append(java.Warngs, errorReference)
			continue
		} else if prefixSplited[0][0:4] == "Log4" { // log4j ignore
			continue
		}

		errorReference = nil
		timeMoment, err := time.ParseInLocation(time.TimeOnly, prefixSplited[0], time.Local)
		if err != nil {
			return err
		}
		currentTime = currentTime.Add(time.Hour*time.Duration(timeMoment.Hour()) + time.Minute*time.Duration(timeMoment.Minute()) + time.Second*time.Duration(timeMoment.Second()))

		if DoneMatch.Match([]byte(prefixSplited[2])) {
			java.ServerPlaform.Started = currentTime
			continue
		}

		contentExplode := js_types.Slice[string](strings.Fields(prefixSplited[2]))
		switch contentExplode.At(0) {
		case "RCON":
			if contentExplode.At(-2) == "on" {
				addr, err := netip.ParseAddrPort(contentExplode.At(-1))
				if err != nil {
					return err
				}
				java.ServerPlaform.Ports = append(java.ServerPlaform.Ports, &logs.Port{AddrPort: addr, From: "RCON"})
			}
		case "Starting":
			switch contentExplode.At(-2) {
			case "version":
				version := contentExplode.At(-1)
				java.ServerPlaform.Version = version
			case "on":
				Value := contentExplode.At(-1)
				if Value[0] == '*' {
					Value = Value[2:]
				}
				port, _ := strconv.ParseInt(Value, 10, 16)
				java.ServerPlaform.Ports = append(java.ServerPlaform.Ports, &logs.Port{AddrPort: netip.AddrPortFrom(netip.IPv4Unspecified(), uint16(port)), From: "TCP"})
			}
		default:
			switch contentExplode.At(-1) {
			case "game":
				at3 := strings.ToLower(contentExplode.At(-3))
				if slices.Contains([]string{"left", "joined"}, at3) {
					playerName := prefixSplited[2][:strings.LastIndex(prefixSplited[2], contentExplode.At(-3))-1]
					if _, ok := java.Players[playerName]; !ok {
						java.Players[playerName] = []logs.Player{}
					}

					action := logs.Action(0)
					switch at3 {
					case "joined":
						action = logs.Connect
					case "left":
						action = logs.Disconnect
					}

					// Append to struct
					java.Players[playerName] = append(java.Players[playerName], JavaPlayer{
						Username: playerName,
						Actioned: action,
						Timed:    currentTime,
					})
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	} else if valid { // return nil if platform is valid
		return nil
	}
	return logs.ErrSkipPlatform
}
