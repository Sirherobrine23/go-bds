package host

import "testing"

func TestHostVersion(t *testing.T) {
	version := HostVersion()
	if len(version) == 0 {
		t.Logf("cannot get host java version")
		return
	}

	t.Log(version)
}
