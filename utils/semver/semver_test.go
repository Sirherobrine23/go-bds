package semver

import (
	"slices"
	"testing"
)

func TestSemver(t *testing.T) {
	versionsToSort, versionsSpect := []VersionString{"0.1.3", "0.1.2", "0.1.0", "0.2.0", "0.1.1", "0.3.0"}, []VersionString{"0.1.0", "0.1.1", "0.1.2", "0.1.3", "0.2.0", "0.3.0"}
	Sort(versionsToSort)
	if !slices.Equal(versionsSpect, versionsToSort) {
		t.Errorf("versions not same sort:\nRequire:  %+v\nReturned: %+v", versionsSpect, versionsToSort)
	}
}
