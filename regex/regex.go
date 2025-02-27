package regex

import regex "regexp"

type Regexp struct{ *regex.Regexp }

func MustCompile(str string) *Regexp {
	return &Regexp{Regexp: regex.MustCompile(str)}
}

func MustCompilePOSIX(str string) *Regexp {
	return &Regexp{Regexp: regex.MustCompilePOSIX(str)}
}

func Compile(expr string) (*Regexp, error) {
	ok, err := regex.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &Regexp{Regexp: ok}, nil
}

func CompilePOSIX(expr string) (*Regexp, error) {
	ok, err := regex.CompilePOSIX(expr)
	if err != nil {
		return nil, err
	}
	return &Regexp{Regexp: ok}, nil
}

// FindAllGroups returns a map with each match group. The map key corresponds to the match group name.
// A nil return value indicates no matches.
func (re *Regexp) FindAllGroups(s string) map[string]string {
	matches := re.FindStringSubmatch(s)
	subnames := re.SubexpNames()
	if matches == nil || len(matches) != len(subnames) {
		return nil
	}

	matchMap := map[string]string{}
	for i := 1; i < len(matches); i++ {
		matchMap[subnames[i]] = matches[i]
	}
	return matchMap
}
