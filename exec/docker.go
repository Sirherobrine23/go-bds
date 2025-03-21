package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/netip"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var _ Proc = &DockerContainer{}

type DockerContainer struct {
	DockerClient *client.Client

	ContainerName string
	Image         string
	Platform      string // Docker platform to run image
	Ports         []nat.Port
	Volumes       []string

	containerID string
	statusExit  *container.WaitResponse
}

// Return new docker exec
func NewDocker(client *client.Client) *DockerContainer {
	return &DockerContainer{
		DockerClient: client,
		Image:        "debian:latest",
		Volumes:      []string{},
		Ports:        []nat.Port{},
	}
}

// Create Docker client connection and return new DockerContainer with "debian:latest" image
func NewDockerDefault() (*DockerContainer, error) {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return NewDocker(client), nil
}

// Append port on start conteiner
func (docker *DockerContainer) AddPort(network string, local, remote uint16) {
	switch network {
	case "udp", "udp4", "udp6":
		docker.Ports = append(docker.Ports, nat.Port(fmt.Sprintf("%d:%d/udp", remote, local)))
	case "tcp", "tcp4", "tcp6":
		docker.Ports = append(docker.Ports, nat.Port(fmt.Sprintf("%d:%d/tcp", remote, local)))
	default:
		docker.Ports = append(docker.Ports, nat.Port(fmt.Sprintf("%d:%d", remote, local)))
	}
}

// Get container addresses
func (docker DockerContainer) ContainerAddr() ([]netip.Addr, error) {
	if docker.containerID == "" {
		return nil, ErrNoRunning
	}

	// Get container info
	info, err := docker.DockerClient.ContainerInspect(context.Background(), docker.containerID)
	if err != nil {
		return nil, err
	}

	// IP Addresses
	addr := []netip.Addr{}
	if info.NetworkSettings != nil {
		if network := info.NetworkSettings; len(info.NetworkSettings.Networks) == 0 {
			if network.IPAddress != "" {
				ipv4Addr, err := netip.ParseAddr(network.IPAddress)
				if err != nil {
					return nil, err
				}
				addr = append(addr, ipv4Addr)
			}

			if network.GlobalIPv6Address != "" {
				ipv6Addr, err := netip.ParseAddr(network.GlobalIPv6Address)
				if err != nil {
					return nil, err
				}
				addr = append(addr, ipv6Addr)
			}
		}

		for _, network := range info.NetworkSettings.Networks {
			if network.IPAddress != "" {
				ipv4Addr, err := netip.ParseAddr(network.IPAddress)
				if err != nil {
					return nil, err
				}
				addr = append(addr, ipv4Addr)
			}

			if network.GlobalIPv6Address != "" {
				ipv6Addr, err := netip.ParseAddr(network.GlobalIPv6Address)
				if err != nil {
					return nil, err
				}
				addr = append(addr, ipv6Addr)
			}
		}
	}

	return addr, nil
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

func (docker *DockerContainer) ExitCode() (int, error) {
	if docker.statusExit != nil {
		return int(docker.statusExit.StatusCode), nil
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
	return int(docker.statusExit.StatusCode), nil
}

func (docker *DockerContainer) Close() error {
	if docker.containerID == "" {
		return ErrNoRunning
	} else if err := docker.DockerClient.ContainerStop(context.Background(), docker.containerID, container.StopOptions{Signal: "SIGTERM"}); err != nil {
		return err
	}
	return docker.Wait()
}

func (docker *DockerContainer) Signal(signal os.Signal) error {
	if docker.containerID == "" {
		return ErrNoRunning
	} else if err := docker.DockerClient.ContainerStop(context.Background(), docker.containerID, container.StopOptions{Signal: signal.String()}); err != nil {
		return err
	}
	return nil
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

	containerCreated, err := docker.DockerClient.ContainerCreate(ctx, &config, &host, nil, nil, docker.ContainerName)
	if docker.containerID = containerCreated.ID; err != nil {
		return err
	} else if err = docker.DockerClient.ContainerStart(ctx, docker.containerID, container.StartOptions{}); err != nil {
		return err
	}

	return nil
}

func (docker *DockerContainer) AppendToStdout(w io.Writer) error { return fs.ErrInvalid }
func (docker *DockerContainer) AppendToStderr(w io.Writer) error { return fs.ErrInvalid }
func (docker *DockerContainer) AppendToStdin(r io.Reader) error  { return fs.ErrInvalid }
