package java_test

import (
	"testing"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java"
)

func TestAll(t *testing.T) {
	distsVersions, err := java.AllReleases()
	if err != nil {
		t.Error(err)
		return
	}

	if len(distsVersions) < 4 {
		t.Errorf("cannot get all distros versions")
		return
	}
}
