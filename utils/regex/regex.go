package regex

import (
	"iter"
	"maps"
	regex "regexp"
)

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

// FindAllGroupSeq returns a map with each match group. The map key corresponds to the match group name.
// A nil return value indicates no matches.
func (re *Regexp) FindAllGroupSeq(s string) iter.Seq2[string, string] {
	matches := re.FindStringSubmatch(s)
	subnames := re.SubexpNames()
	if matches == nil || subnames == nil || len(matches) != len(subnames) {
		return func(yield func(string, string) bool) {}
	}
	return func(yield func(string, string) bool) {
		for i := range len(matches) {
			if subnames[i] != "" {
				if !yield(subnames[i], matches[i]) {
					return
				}
			}
		}
	}
}

// FindAllGroup returns a map with each match group. The map key corresponds to the match group name.
// A nil return value indicates no matches.
func (re *Regexp) FindAllGroup(s string) map[string]string {
	return maps.Collect(re.FindAllGroupSeq(s))
}

func (re *Regexp) AllStringGroupSeq(s string) iter.Seq[iter.Seq2[string, string]] {
	subnames := re.SubexpNames()
	if subnames == nil {
		return func(yield func(iter.Seq2[string, string]) bool) {}
	}
	return func(yield func(iter.Seq2[string, string]) bool) {
		for _, matches := range re.FindAllStringSubmatch(s, -1) {
			if len(matches) != len(subnames) {
				continue
			}

			done := yield(func(yield func(string, string) bool) {
				for i := range len(matches) {
					if subnames[i] != "" {
						if !yield(subnames[i], matches[i]) {
							return
						}
					}
				}
			})
			if !done {
				break
			}
		}
	}
}

func (re *Regexp) AllStringGroups(s string) []map[string]string {
	n1 := []map[string]string{}
	for values := range re.AllStringGroupSeq(s) {
		n1 = append(n1, maps.Collect(values))
	}
	return n1
}
