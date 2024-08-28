package mojang

import "testing"

func TestVersions(t *testing.T) {
	if _, err := FromVersions(); err != nil {
		t.Error(err)
		return
	}
}
