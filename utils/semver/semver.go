package semver

import (
	"encoding"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/utils/regex"
)

var (
	semverMatch   = regex.MustCompile(`(?m)^[a-z]?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)(\.(?P<patch>0|[1-9]\d*))?([\._](?P<extra>[0-9\.]+))?(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?[a-z]?$`)
	mojangVersion = regex.MustCompile(`(?m)^(?P<major>[0-9]\d*)(w(?P<minor>[0-9]+)(?P<extra>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-_]*))?)$`)
)

// Return [Version] from type
type Semver interface {
	SemverVersion() Version
}

// Return version string from type
type Versioner interface {
	Version() string
}

// Semver version
type Version interface {
	fmt.Stringer
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	Compare(Version) int
	LessThan(Version) bool
	Equal(Version) bool
}

// Return [Version] from string, [Version], [Semver], [Versioner], [fmt.Stringer] and [encoding.TextMarshaler]
func ExtractVersion(d any) Version {
	switch v := d.(type) {
	case Version:
		return v
	case Semver:
		return v.SemverVersion()
	case Versioner:
		return New(v.Version())
	case fmt.Stringer:
		return New(v.String())
	case string:
		return New(v)
	case encoding.TextMarshaler:
		text, err := v.MarshalText()
		if err != nil {
			return nil
		}
		return New(string(text))
	}
	return nil
}

type sortVersions[T any] []T

func (s sortVersions[_]) Len() int      { return len(s) }
func (s sortVersions[_]) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortVersions[_]) Less(i, j int) bool {
	si, sj := ExtractVersion(s[i]), ExtractVersion(s[j])
	if si == nil || sj == nil {
		return false
	}
	return si.LessThan(sj)
}

type VersionString string

func (ver VersionString) Version() string { return string(ver) }

// Sort sorts the given slice of Version
func Sort[T any](versions []T) {
	sort.Sort(sortVersions[T](versions))
}

func New(version string) Version {
	switch version {
	case "":
		return nil
	case "latest", "nightly":
		return &versionInfo{Original: version, Major: 9999, Minor: 9999, Patch: 9999}
	}

	// Parse the version string into major, minor, patch, pre-release, and build components
	versionStr := &versionInfo{}
	if err := versionStr.UnmarshalText([]byte(strings.Join(strings.Fields(version), "-"))); err != nil {
		return nil
	}
	return versionStr
}

var _ Version = &versionInfo{}

type versionInfo struct {
	Original string
	Major    int
	Minor    int
	Patch    int
	Extras   []int
	Pre      string
	Build    string
}

func (v *versionInfo) String() string {
	return v.Original
}

func (v *versionInfo) MarshalText() ([]byte, error) {
	return []byte(v.Original), nil
}
func (v *versionInfo) Compare(other Version) int {
	verInt := func(n int) int {
		if n == 0 {
			return 0
		} else if n < 0 {
			return -1
		}
		return 1
	}
	if ov, ok := other.(*versionInfo); ok && ov != v {
		if v.Major != ov.Major {
			return verInt(v.Major - ov.Major)
		} else if v.Minor != ov.Minor {
			return verInt(v.Minor - ov.Minor)
		} else if v.Patch != ov.Patch {
			return verInt(v.Patch - ov.Patch)
		}
		for i := 0; i < len(v.Extras) && i < len(ov.Extras); i++ {
			if v.Extras[i] != ov.Extras[i] {
				return verInt(v.Extras[i] - ov.Extras[i])
			}
		}
		if v.Pre != ov.Pre {
			if v.Pre == "" {
				return 1
			} else if ov.Pre == "" {
				return -1
			}
			return verInt(len(v.Pre) - len(ov.Pre))
		} else if v.Build != ov.Build {
			if v.Build == "" {
				return 1
			} else if ov.Build == "" {
				return -1
			}
			return verInt(len(v.Build) - len(ov.Build))
		}
	}
	return 0
}

func (v *versionInfo) LessThan(other Version) bool {
	return v.Compare(other) < 0
}

func (v *versionInfo) Equal(other Version) bool {
	return v.Compare(other) == 0
}

func (v *versionInfo) UnmarshalText(data []byte) error {
	version := string(data)
	if len(version) == 0 {
		return nil
	}

	mustInt := func(v string) int {
		n, _ := strconv.Atoi(v)
		return n
	}

	switch {
	default:
		return errors.New("invalid semver version")
	case semverMatch.MatchString(version):
		if groups := semverMatch.FindAllGroups(version); groups != nil {
			v.Major = mustInt(groups["major"])
			v.Minor = mustInt(groups["minor"])
			v.Patch = mustInt(groups["patch"])
			v.Pre = groups["prerelease"]
			v.Build = groups["buildmetadata"]
			if x := groups["extra"]; x != "" {
				for n := range strings.SplitSeq(strings.Trim(x, "."), ".") {
					v.Extras = append(v.Extras, mustInt(n))
				}
			}
		}
	case mojangVersion.MatchString(version):
		if groups := semverMatch.FindAllGroups(version); groups != nil {
			if minor2, ok := groups["minor2"]; ok {
				v.Major = mustInt(groups["major"])
				v.Minor = mustInt(minor2)
				v.Patch = mustInt(groups["patch"])
				v.Build = groups["extra2"]
			} else {
				v.Major = 0
				v.Minor = mustInt(groups["major"])
				v.Patch = mustInt(groups["minor"])
				v.Build = groups["extra"]
			}
		}
	}
	v.Original = string(version)
	return nil
}
