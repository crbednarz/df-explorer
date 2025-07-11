package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/docker/docker/client"
)

type Server struct {
	socketPath string
	listener   net.Listener
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
	socketPath := "/tmp/.df-explorer.sock"
	os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	os.Chmod(socketPath, 0666)
	server := &Server{
		socketPath: socketPath,
		listener:   listener,
	}
	return server, nil
}

func (s *Server) SpawnContainer(ctx context.Context, cli *client.Client, image string) (*docker.Container, error) {
	container, err := docker.NewContainer(ctx, cli, image, docker.WithMount(
		s.socketPath,
		s.socketPath,
	))
	return container, err
}

func (s *Server) Listen(callback CommandCallback) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var request commandJSON
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		command := Command{
			Command: request.Command,
		}
		err = callback(command)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		response := responseJSON{
			State:   "success",
			Message: fmt.Sprintf("Command '%s' executed successfully", command.Command),
		}
		json.NewEncoder(w).Encode(response)
	})

	httpServer := http.Server{
		Handler: mux,
	}
	return httpServer.Serve(s.listener)
}

func (s *Server) Close() error {
	s.listener.Close()
	return os.Remove(s.socketPath)
}
