package explorer

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/nxadm/tail"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/docker/docker/client"
)

type Server struct {
	sessionPath   string
	remoteLogPath string
}

type commandJSON struct {
	Command string `json:"command"`
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

	server := &Server{
		sessionPath:   sessionDir,
		remoteLogPath: "/tmp/df-explorer",
	}
	return server, nil
}

func (s *Server) SpawnContainer(ctx context.Context, cli *client.Client, image string) (*docker.Container, error) {
	container, err := docker.NewContainer(ctx, cli, image, docker.WithMount(
		s.sessionPath,
		s.remoteLogPath,
	))
	return container, err
}

func (s *Server) historyLogPath() string {
	return path.Join(s.sessionPath, "history.log")
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
	return os.Remove(s.sessionPath)
}
