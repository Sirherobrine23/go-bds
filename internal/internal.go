package internal

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"regexp"
)

// FindAllGroups returns a map with each match group. The map key corresponds to the match group name.
// A nil return value indicates no matches.
func FindAllGroups(re *regexp.Regexp, s string) map[string]string {
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

func MountFindAllGroups(re string, s string) (map[string]string, error) {
	rem, err := regexp.Compile(re)
	if err != nil {
		return nil, err
	}
	return FindAllGroups(rem, s), nil
}

func ArrayStringIncludes(arr []string, names ...string) (string, bool) {
	for _, n := range arr {
		for _, name := range names {
			if n == name {
				return name, true
			}
		}
	}
	return "", false
}

func SHA1(stream io.Reader) string {
	sha1New := sha1.New()
	io.Copy(sha1New, stream)
	return hex.EncodeToString(sha1New.Sum(nil))
}

func SHA256(stream io.Reader) string {
	sha1New := sha256.New()
	io.Copy(sha1New, stream)
	return hex.EncodeToString(sha1New.Sum(nil))
}

func ExistPath(path string) bool {
	_, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}