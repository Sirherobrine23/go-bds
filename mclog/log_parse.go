package mclog

import (
	"bytes"
	"errors"
	"io"
)

var (
	ErrCannotParse error = errors.New("cannot detect server")
	ErrSkipParse   error = errors.New("skip parse log")
)

type ServerParse interface {
	Detect(log io.ReadSeeker) error // Check if log is compatible and parse
	Insight() *Insights             // Return mclog Insights
}

type PlatformParse interface {
	String() string            // Platform name, ex: 'mojang:bedrock', 'mojang:java', 'spigot', 'paper', 'pocketmine'
	New() (ServerParse, error) // Return new server parse
}

var logParses = []PlatformParse{}

func ParsesRegitred() []string {
	names := []string{}
	for _, n := range logParses {
		names = append(names, n.String())
	}
	return names
}

// Add new server log process
func RegisterNewParse(loger PlatformParse) bool {
	if loger == nil {
		return false
	}
	for _, value := range logParses {
		if value == loger || value.String() == loger.String() {
			return false
		}
	}
	logParses = append(logParses, loger)
	return true
}

// Parse log
func ParseLog(input io.Reader) (*Insights, error) {
	if st, ok := input.(io.ReadSeeker); !ok {
		buffer, err := io.ReadAll(st)
		if err != nil {
			return nil, err
		}
		input = bytes.NewReader(buffer)
	}

	reader := input.(io.ReadSeeker)
	for _, server := range logParses {
		parse, err := server.New()
		if err != nil {
			return nil, err
		}

		if err = parse.Detect(reader); err == ErrSkipParse {
			// Reset stream
			if _, err = reader.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			continue // Skip parse
		} else if err != nil {
			return nil, err
		}

		return parse.Insight(), nil
	}
	return nil, ErrCannotParse
}