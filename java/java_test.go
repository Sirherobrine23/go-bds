package java

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test spigot build to install server
func TestSpigotBuild(t *testing.T) {
	// Static Version
	version := SpigotMC{
		MCVersion:    "1.21.1",
		ToolVersion:  181,
		JavaVersions: []uint{65, 67},
		Ref: struct {
			Spigot      string `json:"Spigot"`
			Bukkit      string `json:"Bukkit"`
			CraftBukkit string `json:"CraftBukkit"`
		}{"a759b629cbf86401aab56b8c3f21a635e9e76c15", "bb4e97c60d2978a1d008f21295a5234228341e14", "0a7bd6c81a33cfaaa2f4d2456c6b237792f38fe6"},
	}

	// Build and install server
	if err := version.Install(filepath.Join(t.TempDir(), "spigotBuild")); err != nil {
		t.Error(err)
		return
	}
}

type LogWrite struct {
	T *testing.T
}

func (w LogWrite) Write(p []byte) (int, error) {
	w.T.Log(string(p))
	return len(p), nil
}

// Start oficial server
func TestStartServer(t *testing.T) {
	versions := Versions{}
	if err := versions.FetchMojang(); err != nil {
		t.Skipf("Cannot fetch mojang versions: %s", err)
		return
	}

	// Mount server struct
	tmpFolder := filepath.Join(os.TempDir(), "javaServer")
	serverFolder, javaFolder, cwd := filepath.Join(tmpFolder, "server"), filepath.Join(tmpFolder, "javabin"), filepath.Join(tmpFolder, "cwd")
	for _, folderPath := range []string{serverFolder, javaFolder, cwd} {
		os.MkdirAll(folderPath, 0777)
	}

	serverInfo, err := NewServer(versions[len(versions)-1], serverFolder, javaFolder, cwd)
	if err != nil {
		t.Errorf("Cannot mount or install server: %s", err)
		return
	}

	serverInfo.PID.AppendToStdout(&LogWrite{t})
	serverInfo.PID.AppendToStderr(&LogWrite{t})

	// Start server
	if err := serverInfo.Start(); err != nil {
		t.Errorf("Cannot start server: %s", err)
		return
	}

	// End server
	go func() {
		serverInfo.Stop()
		select {
		case <-time.After(time.Second * 100):
			serverInfo.PID.Kill()
		}
	}()

	// Wait server stop
	if err := serverInfo.PID.Wait(); err != nil {
		t.Errorf("Server status: %s", err)
		return
	}
}

// List versions
func TestVersions(t *testing.T) {
	t.Run("Mojang", func(t *testing.T) {
		vers := Versions{}
		if err := vers.FetchMojang(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Java versions: %s", d)
	})

	t.Run("Spigot", func(t *testing.T) {
		vers := Versions{}
		if err := vers.FetchSpigotVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Spigot versions: %s", d)
	})

	t.Run("Purpur", func(t *testing.T) {
		vers := Versions{}
		if err := vers.FetchPurpurVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Purpur versions: %s", d)
	})

	t.Run("Paper", func(t *testing.T) {
		vers := Versions{}
		if err := vers.FetchPaperVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Paper versions: %s", d)
	})

	t.Run("Folia", func(t *testing.T) {
		vers := Versions{}
		if err := vers.FetchFoliaVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Folia versions: %s", d)
	})

	t.Run("Velocity", func(t *testing.T) {
		vers := Versions{}
		if err := vers.FetchVelocityVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Velocity versions: %s", d)
	})
}
