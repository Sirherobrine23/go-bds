package logs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"reflect"
	"strings"
	"time"
)

var (
	ErrPlayerNotExist       error = errors.New("player not exists")          // Player not exists or not never connected in server session
	ErrSkipPlatform         error = errors.New("skip platform parse")        // Skip current platform parse and continue to next if avaible
	ErrCannotDetectPlatform error = errors.New("cannot detect log platform") // Platform log not detected or not loaded

	reflectParse = map[string]reflect.Type{}
)

// Register parse to log Dectect platform
func RegisterParse[Parse Log](name string) bool {
	if _, ok := reflectParse[name]; !ok {
		reflectParse[name] = reflect.TypeFor[Parse]()
	}
	return true
}

type Action int // Player actions

const (
	_          Action = iota
	Disconnect        // Player disconnected from server
	Connect           // Player connected to server
	Spawned           // Player connected to server and spawned in world server
	Banned            // Player is banned from server by operator

	PlayerActionsValid = Disconnect | Connect | Spawned | Banned // Check if actions is valid
)

func (act Action) String() string {
	switch act {
	case Disconnect:
		return "disconnect"
	case Connect:
		return "connect"
	case Spawned:
		return "spawned"
	case Banned:
		return "banned"
	default:
		return "unknown"
	}
}

func (act Action) MarshalText() ([]byte, error) {
	return []byte(act.String()), nil
}

func (act *Action) UnmarshalText(data []byte) error {
	switch string(data) {
	case "disconnect":
		*act = Disconnect
	case "connect":
		*act = Connect
	case "spawned":
		*act = Spawned
	case "banned":
		*act = Banned
	default:
		return fmt.Errorf("unknown action: %s", data)
	}
	return nil
}

type Port struct {
	AddrPort netip.AddrPort `json:"addr"` // Port and ip address
	From     string         `json:"from"` // Server listened protocol, TCP/UDP/Unix...
}

var _ error = &ErrorReference{}

type ErrorReference struct {
	LogLevel int
	FistLine string
	Line     []string
}

func (err *ErrorReference) Error() string {
	errLevel := "error"
	if err.LogLevel == 1 {
		errLevel = "warning"
	} else if err.LogLevel >= 2 {
		errLevel = "fatal"
	}
	return strings.TrimSpace(fmt.Sprintf("%s: %s", errLevel, err.FistLine))
}

func (err *ErrorReference) MarshalText() ([]byte, error) {
	return []byte(err.Error()), nil
}

func (err *ErrorReference) Unwrap() []error {
	return []error{
		err,
		errors.New(strings.Join(err.Line, "\n")),
	}
}

// Server info
type Server struct {
	Version  string    `json:"version"`  // Server version if avaible
	Started  time.Time `json:"started"`  // Time server started or avaible to connect
	Platform string    `json:"platform"` // Server platform, example: bedrock, java, pocketmine...
	Ports    []*Port   `json:"ports"`    // Server ports listened
}

type Player interface {
	Name() string    // Player username
	Action() Action  // Player action
	Time() time.Time // Action time
	XUID() int64     // Xbox XUID
}

// Implements log parse and return based log info
type Log interface {
	ParseTime(time.Time, io.Reader) error // Parse log with server date
	Parse(io.Reader) error                // Parse log with current time
	Server() *Server                      // Server info
	GetPlayer(string) ([]Player, bool)    // Get player info
	Errors() []error                      // Get log errors
	Warnings() []error                    // Get log warnings
}

// Parse log
func Parse(log io.ReadSeeker) (Log, error) {
	for _, typeOf := range reflectParse {
		platformLog := reflect.New(typeOf.Elem()).Interface().(Log)
		if err := platformLog.Parse(log); err != nil {
			if err == ErrSkipPlatform {
				if _, err = log.Seek(0, io.SeekStart); err == nil {
					continue // Skip
				}
				err = fmt.Errorf("cannot seek log file: %s", err)
			}
			return nil, err
		}
		return platformLog, nil
	}
	return nil, ErrCannotDetectPlatform
}

// Parse string log
func ParseString(log string) (Log, error) {
	return Parse(strings.NewReader(log))
}

// Parse buffer log
func ParseBuffer(log []byte) (Log, error) {
	return Parse(bytes.NewReader(log))
}
