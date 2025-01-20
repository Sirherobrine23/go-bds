package java

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/mclog"
)

var _ = mclog.RegisterNewParse(NewLogParse{})

type NewLogParse struct{}

func (NewLogParse) String() string                  { return "mojang/java" }
func (NewLogParse) New() (mclog.ServerParse, error) { return &LogerParse{}, nil }

type LogerParse struct {
	local mclog.Insights
}

func (loger LogerParse) Insight() *mclog.Insights { return &loger.local }
func (loger *LogerParse) Detect(log io.ReadSeeker) error {
	loger.local = mclog.Insights{
		ID:       "mojang/java",
		Name:     "java",
		Type:     "Server log",
		Title:    "java",
		Version:  "unknown",
		Analysis: map[mclog.LogLevel][]*mclog.InsightsAnalysis{},
	}

	scan := bufio.NewScanner(log)
	isValid, Analysis, logLevel, err := false, (*mclog.InsightsAnalysis)(nil), mclog.LogUnknown, error(nil)
	for scan.Scan() {
		add := false
		text := scan.Text()
		if strings.HasPrefix(text, "Unpacking") || strings.HasPrefix(text, "Starting") {
			continue
		} else if !isValid {
			if text[0] != '[' {
				return mclog.ErrSkipParse
			}
			isValid = true
		}
		if text[0] != '[' {
			if Analysis == nil {
				continue
			}
			Analysis.Entry.Lines = append(Analysis.Entry.Lines, mclog.EntryLine{
				Content: text,
				Numbers: strings.Count(text, "") - len(strings.Split(text, " ")),
			})
			continue
		}
		Analysis = &mclog.InsightsAnalysis{
			Value: text,
			Entry: mclog.AnalysisEntry{
				Lines: []mclog.EntryLine{
					{
						Content: text,
						Numbers: strings.Count(text, "") - len(strings.Split(text, " ")),
					},
				},
			},
		}

		prefixSplited := [3]string(strings.SplitAfterN(text, "]", 3))
		prefixSplited[0] = strings.Replace(strings.Replace(strings.TrimSpace(prefixSplited[0][1:]), "[", "", 1), "]", "", 1)
		prefixSplited[1] = strings.Replace(strings.Replace(strings.TrimSpace(prefixSplited[1][1:]), "[", "", 1), "]", "", 1)
		prefixSplited[2] = strings.TrimSpace(prefixSplited[2][1:])
		Analysis.Message = prefixSplited[2]
		Analysis.Entry.Prefix = fmt.Sprintf("[%s] [%s]", prefixSplited[0], prefixSplited[1])
		if len(strings.Split(prefixSplited[0], ":")) != 3 {
			continue
		}

		if strings.HasSuffix(prefixSplited[1], "INFO") {
			logLevel = mclog.LogInfo
		} else if strings.HasSuffix(prefixSplited[1], "WARN") {
			logLevel = mclog.LogWarn
			Analysis.Label = "Error"
			add = true
		}

		if Analysis.Entry.EntryTime, err = time.ParseInLocation(time.TimeOnly, prefixSplited[0], time.Local); err != nil {
			return err
		}
		Analysis.Entry.EntryTime = Analysis.Entry.EntryTime.UTC()
		contentExplode := strings.Fields(prefixSplited[2])

		switch contentExplode[0] {
		case "RCON":
			if contentExplode[len(contentExplode)-2] == "on" {
				add = true
				Analysis.Label = "RCON"
				Analysis.Value = contentExplode[len(contentExplode)-1]
			}
		case "Starting":
			if len(contentExplode) <= 3 {
				continue
			}

			switch contentExplode[len(contentExplode)-2] {
			case "version":
				add = true
				version := contentExplode[len(contentExplode)-1]
				loger.local.Version = version
				Analysis.Value = version
				Analysis.Label = "Java version"
			case "on":
				add = true
				Analysis.Label = "Port"
				Analysis.Value = contentExplode[len(contentExplode)-1][2:]
			}
		default:
			switch contentExplode[len(contentExplode)-1] {
			case "game":
				if slices.Contains([]string{"left", "joined"}, contentExplode[len(contentExplode)-3]) {
					add = true
					Analysis.Message = "Player status"
					Analysis.Label = contentExplode[len(contentExplode)-3]
					Analysis.Value = prefixSplited[2][:strings.LastIndex(prefixSplited[2], contentExplode[len(contentExplode)-3])-1]
				}
			}
		}

		if add {
			Analysis.Counter = len(Analysis.Entry.Lines)
			loger.local.Analysis[logLevel] = append(loger.local.Analysis[logLevel], Analysis)
		}
	}
	return nil
}
