package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

var (
	ErrNoContainerId  = errors.New("set container ID")
	ErrNoDockerClient = errors.New("set docker client")
)

type Docker struct {
	DockerClient      *client.Client    `json:"-"`                 // Docker client
	DockerImage       string            `json:"dockerImage"`       // Docker image, default is docker.io/ubuntu:latest
	Network           string            `json:"network"`           // Network host to container run
	Ports             map[string]string `json:"ports"`             // Expose ports: map[string]string{"22":"8022", "80": ""}
	Volumes           []string          `json:"volumes"`           // Volumes to mount, exp: []string{`/var/lib/docker`}
	LocalFolders      []string          `json:"folders"`           // Local folder to mount in container, exp: []string{`./data:/var/lib/data`}
	ReplaceEntrypoint bool              `json:"replaceEntrypoint"` // Replace default entrypoint to /bin/sh
	ContainerID       string            `json:"-"`                 // Container ID

	codeExit *int64 // Container exit code
}

func (w Docker) ExitCode() (int64, error) {
	if w.codeExit == nil {
		return 0, ErrRunning
	}
	return *w.codeExit, nil
}

func (w Docker) Kill() error {
	if w.ContainerID == "" {
		return ErrNoContainerId
	}
	return w.DockerClient.ContainerStop(context.Background(), w.ContainerID, container.StopOptions{Signal: "SIGKILL"})
}

func (w Docker) Close() error {
	if w.ContainerID == "" {
		return ErrNoContainerId
	}
	return w.DockerClient.ContainerStop(context.Background(), w.ContainerID, container.StopOptions{Signal: "SIGTERM"})
}

func (w Docker) Wait() error {
	notChannel, errNotChannel := w.DockerClient.ContainerWait(context.Background(), w.ContainerID, container.WaitConditionNotRunning)
	exitChannel, errExitChannel := w.DockerClient.ContainerWait(context.Background(), w.ContainerID, container.WaitConditionNextExit)
	deletedChannel, errDeletedChannel := w.DockerClient.ContainerWait(context.Background(), w.ContainerID, container.WaitConditionRemoved)

	var status container.WaitResponse
	var err error
	select {
	case err = <- errDeletedChannel:
	case err = <- errExitChannel:
	case err = <- errNotChannel:
	case status = <- deletedChannel:
	case status = <- exitChannel:
	case status = <- notChannel:
	}
	
	if err != nil {
		return err
	} else if status.Error != nil {
		return errors.New(status.Error.Message)
	} else if status.StatusCode != 0 {
		return fmt.Errorf("exit code %d", status.StatusCode)
	}
	return nil
}

func (w Docker) Write(p []byte) (int, error) {
	stdin, err := w.StdinFork()
	if err != nil {
		return 0, err
	}
	defer stdin.Close()
	return stdin.Write(p)
}

func (w Docker) StdinFork() (io.WriteCloser, error) {
	if w.ContainerID == "" {
		return nil, ErrNoContainerId
	}
	res, err := w.DockerClient.ContainerAttach(context.Background(), w.ContainerID, container.AttachOptions{Stdin: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return res.Conn, nil
}

func (w *Docker) forkLog(stdout, stderr io.Writer) error {
	res, err := w.DockerClient.ContainerAttach(context.Background(), w.ContainerID, container.AttachOptions{Stdout: true, Stream: true})
	if err != nil {
		return err
	}
	if media, _ := res.MediaType(); media == "application/vnd.docker.multiplexed-stream" {
		go stdcopy.StdCopy(stdout, stderr, res.Reader)
		return nil
	}
	if stdout != io.Discard {
		go io.Copy(stdout, res.Reader)
	}
	if stderr != io.Discard {
		go io.Copy(stderr, res.Reader)
	}
	return nil
}

func (cli Docker) StdoutFork() (io.ReadCloser, error) {
	if cli.ContainerID == "" {
		return nil, ErrNoContainerId
	}
	r, w := io.Pipe()
	if err := cli.forkLog(w, io.Discard); err != nil {
		w.Close()
		return nil, err
	}
	return r, nil
}

func (cli Docker) StderrFork() (io.ReadCloser, error) {
	if cli.ContainerID == "" {
		return nil, ErrNoContainerId
	}
	r, w := io.Pipe()
	if err := cli.forkLog(io.Discard, w); err != nil {
		w.Close()
		return nil, err
	}
	return r, nil
}

func (w *Docker) Start(options ProcExec) error {
	if w.DockerClient == nil {
		return ErrNoDockerClient
	} else if w.DockerImage == "" {
		w.DockerImage = "docker.io/ubuntu:latest"
	}

	ctx := context.Background()
	reader, err := w.DockerClient.ImagePull(ctx, w.DockerImage, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)

	var hostConfig container.HostConfig
	var config container.Config = container.Config{
		Tty:       false,
		Image:     w.DockerImage,
		Cmd:       options.Arguments,
		OpenStdin: true,
		Env:       []string{},
		Volumes:   make(map[string]struct{}),
	}

	if options.Cwd != "" {
		config.WorkingDir = options.Cwd
	}

	if w.ReplaceEntrypoint {
		config.Entrypoint = strslice.StrSlice{"/bin/sh"}
	}

	for key, value := range options.Environment {
		config.Env = append(config.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if w.Network != "" {
		hostConfig.NetworkMode = container.NetworkMode(w.Network)
	}

	hostConfig.PortBindings = nat.PortMap{}
	for p, localPort := range w.Ports {
		hostConfig.PortBindings[nat.Port(p)] = []nat.PortBinding{}
		if localPort != "" {
			hostConfig.PortBindings[nat.Port(p)] = []nat.PortBinding{{HostPort: localPort}}
		}
	}

	for _, vol := range w.Volumes {
		config.Volumes[vol] = struct{}{}
	}

	for _, f := range w.LocalFolders {
		sk := strings.SplitN(f, ":", 2)
		if len(sk) == 2 && sk[0][0] != '.' {
			hostConfig.Binds = append(hostConfig.Binds, f)
		}
	}

	containerInfo, err := w.DockerClient.ContainerCreate(ctx, &config, &hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	w.ContainerID = containerInfo.ID

	if err := w.DockerClient.ContainerStart(ctx, containerInfo.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := w.DockerClient.ContainerWait(ctx, containerInfo.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	go func() {
		statusCh, errCh := w.DockerClient.ContainerWait(ctx, containerInfo.ID, container.WaitConditionNextExit)
		select {
		case err := <-errCh:
			if err != nil {
				return
			}
		case s := <-statusCh:
			w.codeExit = new(int64) // Allocate space to int64
			*w.codeExit = s.StatusCode
		}
	}()

	return nil
}
