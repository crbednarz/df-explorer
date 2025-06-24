package docker

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/appdefaults"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/pkg/errors"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/sync/errgroup"
)

type Builder struct {
	client *client.Client
}

func NewBuilder(ctx context.Context) (*Builder, error) {
	c, err := client.New(ctx, appdefaults.Address)
	if err != nil {
		return nil, err
	}

	return &Builder{
		client: c,
	}, nil
}

func (b *Builder) Commit(ctx context.Context, command string) error {
	pipeR, pipeW := io.Pipe()
	solveOpt, err := newSolveOpt(pipeW)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)
	ch := make(chan *client.SolveStatus)
	eg.Go(func() error {
		_, err := b.client.Solve(ctx, nil, *solveOpt, ch)
		return err
	})
	eg.Go(func() error {
		d, err := progressui.NewDisplay(os.Stderr, progressui.TtyMode)
		if err != nil {
			log.Printf("failed to create progress display: %v", err)
			// If an error occurs while attempting to create the tty display,
			// fallback to using plain mode on stdout (in contrast to stderr).
			d, _ = progressui.NewDisplay(os.Stdout, progressui.PlainMode)
		}
		// not using shared context to not disrupt display but let is finish reporting errors
		_, err = d.UpdateFrom(context.TODO(), ch)
		return err
	})
	eg.Go(func() error {
		if err := loadDockerTar(pipeR); err != nil {
			return err
		}
		return pipeR.Close()
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func newSolveOpt(w io.WriteCloser) (*client.SolveOpt, error) {
	buildCtx := "."
	tag := "test-tag"
	file := filepath.Join(buildCtx, "Dockerfile")

	cxtLocalMount, err := fsutil.NewFS(buildCtx)
	if err != nil {
		return nil, errors.New("invalid buildCtx local mount dir")
	}

	dockerfileLocalMount, err := fsutil.NewFS(filepath.Dir(file))
	if err != nil {
		return nil, errors.New("invalid dockerfile local mount dir")
	}

	frontend := "dockerfile.v0" // TODO: use gateway
	frontendAttrs := map[string]string{
		"filename": filepath.Base(file),
	}
	// if target := clicontext.String("target"); target != "" {
	// 	frontendAttrs["target"] = target
	// }
	return &client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type: "docker", // TODO: use containerd image store when it is integrated to Docker
				Attrs: map[string]string{
					"name": tag,
				},
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return w, nil
				},
			},
		},
		LocalMounts: map[string]fsutil.FS{
			"context":    cxtLocalMount,
			"dockerfile": dockerfileLocalMount,
		},
		Frontend:      frontend,
		FrontendAttrs: frontendAttrs,
	}, nil
}

func loadDockerTar(r io.Reader) error {
	// no need to use moby/moby/client here
	cmd := exec.Command("docker", "load")
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
