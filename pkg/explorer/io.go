package explorer

import (
	"io"
	"sync"
)

type dynamicIO struct {
	writer io.Writer
	reader io.Reader
	mu     sync.RWMutex
}

func (d *dynamicIO) SetReaderWriter(r io.Reader, w io.Writer) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.reader = r
	d.writer = w
}

func (d *dynamicIO) Write(p []byte) (n int, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.writer == nil {
		return 0, nil
	}

	return d.writer.Write(p)
}

func (d *dynamicIO) Read(p []byte) (n int, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.reader == nil {
		return 0, nil
	}
	return d.reader.Read(p)
}
