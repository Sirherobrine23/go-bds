package semver

import (
	"slices"
	"testing"
)

func processVersion(p []string) []*Version {
	n1 := []*Version{}
	for _, v := range p {
		n1 = append(n1, New(v))
	}
	return n1
}
func processString(p []*Version) []string {
	n1 := []string{}
	for _, v := range p {
		n1 = append(n1, v.String())
	}
	return n1
}

func TestSemver(t *testing.T) {
	versionsToSort, versionsSpect := processVersion([]string{"0.1.3", "0.1.2", "0.1.0", "0.2.0", "0.1.1", "0.3.0"}), []string{"0.1.0", "0.1.1", "0.1.2", "0.1.3", "0.2.0", "0.3.0"}
	Sort(versionsToSort)
	if !slices.Equal(versionsSpect, processString(versionsToSort)) {
		t.Errorf("versions not same sort")
	}
}
