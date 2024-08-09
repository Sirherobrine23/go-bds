package exec

import (
	"testing"

	"github.com/docker/docker/client"
)

func TestDocker(t *testing.T) {
	var dockerExec = new(Docker)
	var err error
	if dockerExec.DockerClient, err = client.NewClientWithOpts(client.FromEnv); err != nil {
		t.Skip(err)
		return
	}

	dockerExec.DockerImage = "debian:latest"
	defer dockerExec.Close()

	if err := dockerExec.Start(ProcExec{ Arguments: []string{"echo", "hello world"} }); err != nil {
		t.Fatal(err)
		return
	}

	if err := dockerExec.Wait(); err != nil {
		t.Fatal(err)
		return
	}
}
