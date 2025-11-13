package explorer

import (
	"io"
	"sync"

	"github.com/crbednarz/df-explorer/pkg/docker"
)

// ContainerProxy is a proxy for docker.Container to allow the underlying
// container to be swapped out dynamically.
type ContainerProxy struct {
	container docker.Container
	mu        sync.RWMutex
}

// Write writes to the underlying container's attachment.
// If no container is set, it returns io.EOF.
func (c *ContainerProxy) Write(p []byte) (n int, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.container != nil {
		return c.container.Attachment().Write(p)
	}
	return 0, io.EOF
}

// Read reads from the underlying container's attachment.
// If no container is set, it returns io.EOF.
func (c *ContainerProxy) Read(p []byte) (n int, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.container != nil {
		return c.container.Attachment().Read(p)
	}
	return 0, io.EOF
}

// SetContainer sets the underlying container.
func (c *ContainerProxy) SetContainer(container docker.Container) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container = container
}

// Close closes the underlying container and removes the reference to it.
func (c *ContainerProxy) Close() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.container != nil {
		err := c.container.Close()
		c.container = nil
		return err
	}
	return nil
}
