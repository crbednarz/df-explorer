package explorer

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/nxadm/tail"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/docker/docker/client"
)

type Server struct {
	localSessionDir  string
	remoteSessionDir string
}

type commandJSON struct {
	Command string `json:"command"`
	State   string `json:"state"`
}

type responseJSON struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

type CommandState int

const (
	CommandStateSuccess CommandState = iota
	CommandStateError
	CommandStateRunning
)

type Command struct {
	Command string
	State   CommandState
}

type CommandCallback func(Command) error

//go:embed df-env.sh
var envScript string

func newServer() (*Server, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get user cache directory: %w", err)
	}

	cacheDir := path.Join(userCacheDir, "df-explorer")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("unable to create cache directory: %w", err)
	}

	sessionDir, err := os.MkdirTemp(cacheDir, "session-*")
	if err != nil {
		return nil, fmt.Errorf("unable to create session directory: %w", err)
	}
	historyFile, err := os.Create(path.Join(sessionDir, "history.log"))
	if err != nil {
		return nil, fmt.Errorf("unable to create history log file: %w", err)
	}
	historyFile.Close()

	envScriptName := "df-env.sh"
	err = os.WriteFile(path.Join(sessionDir, envScriptName), []byte(envScript), 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to write profile script: %w", err)
	}

	server := &Server{
		localSessionDir:  sessionDir,
		remoteSessionDir: "/tmp/df-explorer",
	}
	return server, nil
}

func (s *Server) SpawnContainer(ctx context.Context, cli *client.Client, image string) (*docker.Container, error) {
	container, err := docker.NewContainer(
		ctx,
		cli,
		image,
		docker.WithMount(
			s.localSessionDir,
			s.remoteSessionDir,
		),
		docker.WithCommand([]string{"/bin/bash"}),
		docker.WithAttach(true),
	)
	fmt.Fprintf(container.Attachment(), "source %s\nreset\n", path.Join(s.remoteSessionDir, "df-env.sh"))
	return container, err
}

func (s *Server) historyLogPath() string {
	return path.Join(s.localSessionDir, "history.log")
}

func (s *Server) Listen(callback CommandCallback) error {
	file, err := tail.TailFile(s.historyLogPath(), tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		return fmt.Errorf("unable to tail log file: %w", err)
	}

	for line := range file.Lines {
		callback(Command{
			Command: line.Text,
			State:   CommandStateRunning,
		})
	}

	return nil
}

func (s *Server) Close() error {
	return os.Remove(s.localSessionDir)
}
