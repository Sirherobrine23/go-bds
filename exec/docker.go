package exec

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerContainer struct {
	DockerClient *client.Client

	Image   string
	Volumes []string
	Ports   []nat.Port

	containerID string
	statusExit  *container.WaitResponse
}

func (docker *DockerContainer) Kill() error {
	if docker.containerID == "" {
		return ErrNoRunning
	}
	return docker.DockerClient.ContainerKill(context.Background(), docker.containerID, "SIGKILL")
}

func (docker *DockerContainer) Wait() error {
	if docker.containerID == "" {
		return ErrNoRunning
	}
	wait, err := docker.DockerClient.ContainerWait(context.Background(), docker.containerID, container.WaitConditionNextExit)
	select {
	case err := <-err:
		if err != nil {
			return err
		}
	case status := <-wait:
		docker.statusExit = &status
		if status.Error != nil {
			return errors.New(status.Error.Message)
		} else if status.StatusCode != 0 {
			return fmt.Errorf("exit code %d", status.StatusCode)
		}
	}

	return nil
}

func (docker *DockerContainer) ExitCode() (int64, error) {
	if docker.statusExit != nil {
		return docker.statusExit.StatusCode, nil
	} else if docker.containerID == "" {
		return -1, ErrNoRunning
	}

	wait, err := docker.DockerClient.ContainerWait(context.Background(), docker.containerID, container.WaitConditionNextExit)
	select {
	case err := <-err:
		if err != nil {
			return -1, err
		}
	case status := <-wait:
		docker.statusExit = &status
		if status.Error != nil {
			return -1, errors.New(status.Error.Message)
		} else if status.StatusCode != 0 {
			return -1, fmt.Errorf("exit code %d", status.StatusCode)
		}
	}
	return docker.statusExit.StatusCode, nil
}

func (docker *DockerContainer) Close() error {
	if docker.containerID == "" {
		return ErrNoRunning
	} else if err := docker.DockerClient.ContainerStop(context.Background(), docker.containerID, container.StopOptions{Signal: "SIGTERM"}); err != nil {
		return err
	}
	return docker.Wait()
}

func (docker *DockerContainer) Write(p []byte) (int, error) {
	pipe, err := docker.StdinFork()
	if err != nil {
		return 0, err
	}
	defer pipe.Close()
	return pipe.Write(p)
}

func (docker *DockerContainer) StdinFork() (io.WriteCloser, error) {
	if docker.containerID == "" {
		return nil, ErrNoRunning
	}
	hj, err := docker.DockerClient.ContainerAttach(context.Background(), docker.containerID, container.AttachOptions{Stdin: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return io.WriteCloser(hj.Conn), nil
}
func (docker *DockerContainer) StdoutFork() (io.ReadCloser, error) {
	if docker.containerID == "" {
		return nil, ErrNoRunning
	}
	return docker.DockerClient.ContainerLogs(context.Background(), docker.containerID, container.LogsOptions{Follow: true, ShowStdout: true})
}
func (docker *DockerContainer) StderrFork() (io.ReadCloser, error) {
	if docker.containerID == "" {
		return nil, ErrNoRunning
	}
	return docker.DockerClient.ContainerLogs(context.Background(), docker.containerID, container.LogsOptions{Follow: true, ShowStderr: true})
}

func (docker *DockerContainer) Start(options ProcExec) error {
	if docker.DockerClient == nil {
		var err error
		if docker.DockerClient, err = client.NewClientWithOpts(client.FromEnv); err != nil {
			return err
		}
	}

	// Return error if container started
	if docker.containerID != "" {
		return ErrRunning
	}

	ctx := context.Background()
	dVolumes, dPorts := map[string]struct{}{}, nat.PortSet{}
	for _, vol := range docker.Volumes {
		dVolumes[vol] = struct{}{}
	}
	for _, port := range docker.Ports {
		dPorts[port] = struct{}{}
	}

	config := container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		Env:          options.Environment.ToSlice(),
		Cmd:          []string{},
		Entrypoint:   options.Arguments,
		Image:        docker.Image,
		WorkingDir:   options.Cwd,
		Volumes:      dVolumes,
		ExposedPorts: dPorts,
	}

	host := container.HostConfig{
		AutoRemove: true,
		Tmpfs: map[string]string{
			"/tmp": "rw,nosuid",
		},
	}

	containerCreated, err := docker.DockerClient.ContainerCreate(ctx, &config, &host, nil, nil, "")
	if docker.containerID = containerCreated.ID; err != nil {
		return err
	} else if err = docker.DockerClient.ContainerStart(ctx, docker.containerID, container.StartOptions{}); err != nil {
		return err
	}

	return nil
}
