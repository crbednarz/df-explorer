package explorer

import (
	"bytes"
	"io"
	"testing"
)

type mockContainer struct {
	data        bytes.Buffer
	ResizeError error
	CloseError  error
	Width       uint
	Height      uint
	IsClosed    bool
}

func newMockContainer() *mockContainer {
	return &mockContainer{IsClosed: false}
}

func newMockContainerWithData(initialData []byte) *mockContainer {
	mc := &mockContainer{IsClosed: false}
	mc.data.Write(initialData)
	return mc
}

func (mc *mockContainer) Attachment() io.ReadWriter {
	return mc
}

func (mc *mockContainer) SetSize(width uint, height uint) error {
	mc.Width = width
	mc.Height = height
	return mc.ResizeError
}

func (mc *mockContainer) Write(p []byte) (n int, err error) {
	return mc.data.Write(p)
}

func (mc *mockContainer) Read(p []byte) (n int, err error) {
	return mc.data.Read(p)
}

func (mc *mockContainer) Close() error {
	mc.IsClosed = true
	return mc.CloseError
}

func TestProxyEmptyWrite(t *testing.T) {
	proxy := &ContainerProxy{}
	data := []byte("hello world")
	// Test writing when no container is set
	n, err := proxy.Write(data)
	if err != nil {
		t.Fatalf("expected io.EOF when writing with no container, got: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 bytes written, got: %d", n)
	}
}

func TestProxyEmptyRead(t *testing.T) {
	proxy := &ContainerProxy{}
	buf := make([]byte, 20)
	// Test reading when no container is set
	n, err := proxy.Read(buf)
	if err != nil {
		t.Fatalf("expected io.EOF when reading with no container, got: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 bytes read, got: %d", n)
	}
}

func TestProxyWrite(t *testing.T) {
	proxy := &ContainerProxy{}
	mock := newMockContainer()
	proxy.SetContainer(mock)
	data := []byte("hello world")
	n, err := proxy.Write(data)
	if err != nil {
		t.Fatalf("unexpected error on write: %v", err)
	}
	if n != len(data) {
		t.Fatalf("expected %d bytes written, got: %d", len(data), n)
	}
	if mock.data.String() != "hello world" {
		t.Fatalf("expected mock container to have 'hello world', got: %s", mock.data.String())
	}
}

func TestProxyRead(t *testing.T) {
	proxy := &ContainerProxy{}
	data := []byte("hello world")
	mock := newMockContainerWithData(data)
	proxy.SetContainer(mock)
	buf := make([]byte, len(data))
	n, err := proxy.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error on read: %v", err)
	}
	if n != len(data) {
		t.Fatalf("expected %d bytes read, got: %d", len(data), n)
	}
	if string(buf) != "hello world" {
		t.Fatalf("expected to read 'hello world', got: %s", string(buf))
	}
}

func TestProxyClose(t *testing.T) {
	proxy := &ContainerProxy{}
	mock := newMockContainer()
	proxy.SetContainer(mock)
	err := proxy.Close()
	if err != nil {
		t.Fatalf("unexpected error on close: %v", err)
	}
	if !mock.IsClosed {
		t.Fatalf("expected mock container to be closed")
	}
}

func TestProxyCloseNoContainer(t *testing.T) {
	proxy := &ContainerProxy{}
	err := proxy.Close()
	if err != nil {
		t.Fatalf("unexpected error on close with no container: %v", err)
	}
}

func TestProxyCloseError(t *testing.T) {
	proxy := &ContainerProxy{}
	mock := newMockContainer()
	mock.CloseError = io.ErrUnexpectedEOF
	proxy.SetContainer(mock)
	err := proxy.Close()

	if err != io.ErrUnexpectedEOF {
		t.Fatalf("expected io.ErrUnexpectedEOF on close, got: %v", err)
	}
}
