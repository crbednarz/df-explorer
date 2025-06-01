package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/jroimartin/gocui"
	"golang.org/x/term"
)

type TerminalView struct {
	pty           *os.File
	oldStdinState term.State
}

func (t *TerminalView) Layout(view *gocui.View) error {
	// oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	// if err != nil {
	// 	panic(err)
	// }
	// defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// go func() { _, _ = io.Copy(t.pty, os.Stdin) }()
	// _, _ = io.Copy(os.Stdout, t.pty)
	return nil
}

func (t *TerminalView) Close() {
	t.pty.Close()
	term.Restore(int(os.Stdin.Fd()), &t.oldStdinState)
}

func NewTerminal(command string) (*TerminalView, error) {
	c := exec.Command("bash")

	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, fmt.Errorf("unable to start terminal: %w", err)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() { _, _ = io.Copy(t.pty, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, t.pty)
	return &TerminalView{
		pty: ptmx,
	}, nil
}

type LayoutManager struct {
	terminal *TerminalView
}

func NewLayoutManager() *LayoutManager {
	return &LayoutManager{
		terminal: nil,
	}
}

func (l *LayoutManager) SetTerminal(terminal *TerminalView) {
	l.terminal = terminal
}

func (l *LayoutManager) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("hello", 0, 1, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return fmt.Errorf("unable to set view: %w", err)
		}

		fmt.Fprintln(v, "Hello world!")
	}
	return nil
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicf("error during gui creation: %v", err)
	}
	defer g.Close()

	terminal, err := NewTerminal("bash")
	if err != nil {
		log.Panicln(err)
	}
	defer terminal.Close()

	layout := NewLayoutManager()
	layout.SetTerminal(terminal)

	g.SetManager(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
