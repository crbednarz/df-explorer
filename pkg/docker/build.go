package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	"github.com/crbednarz/df-explorer/pkg/util"
	dclient "github.com/docker/docker/client"
	bclient "github.com/moby/buildkit/client"
	_ "github.com/moby/buildkit/client/connhelper/dockercontainer"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/sync/errgroup"
)

type Builder struct {
	client *bclient.Client
	daemon *Container
}

type BuildConfig struct {
	BuildContext string
	Dockerfile   string
}

func NewBuilder(ctx context.Context, dockerClient *dclient.Client) (*Builder, error) {
	cacheDir, err := util.CacheDir()
	if err != nil {
		return nil, err
	}
	builderCache := path.Join(cacheDir, "buildkitd")

	daemon, err := NewContainer(
		ctx,
		dockerClient,
		"moby/buildkit:rootless",
		WithName("df-buildkitd"),
		WithSecurityOption("seccomp=unconfined"),
		WithSecurityOption("apparmor=unconfined"),
		WithMount(builderCache, "/var/lib/buildkit"),
		WithPull(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create buildkit daemon container: %w", err)
	}

	c, err := bclient.New(ctx, "docker-container://df-buildkitd")
	if err != nil {
		return nil, err
	}

	return &Builder{
		client: c,
		daemon: daemon,
	}, nil
}

func (b *Builder) Build(ctx context.Context, config BuildConfig) (string, error) {
	pipeR, pipeW := io.Pipe()
	solveOpt, err := newSolveOpt(pipeW, config)
	var imageID string
	if err != nil {
		return "", err
	}
	eg, ctx := errgroup.WithContext(ctx)
	ch := make(chan *bclient.SolveStatus)
	eg.Go(func() error {
		_, err := b.client.Solve(ctx, nil, *solveOpt, ch)
		if err != nil {
			pipeW.CloseWithError(err)
		}
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
		_, err = d.UpdateFrom(ctx, ch)
		return err
	})
	eg.Go(func() error {
		imageID, err = loadDockerTar(pipeR)
		if err != nil {
			return err
		}
		return pipeR.Close()
	})
	if err := eg.Wait(); err != nil {
		return "", err
	}
	return imageID, nil
}

func newSolveOpt(w io.WriteCloser, config BuildConfig) (*bclient.SolveOpt, error) {
	cxtLocalMount, err := fsutil.NewFS(config.BuildContext)
	if err != nil {
		return nil, fmt.Errorf("invalid build context dir (%s): %w", config.BuildContext, err)
	}

	dockerfileLocalMount, err := fsutil.NewFS(filepath.Dir(config.Dockerfile))
	if err != nil {
		return nil, fmt.Errorf("invalid dockerfile dir (%s): %w", config.Dockerfile, err)
	}

	return &bclient.SolveOpt{
		Exports: []bclient.ExportEntry{
			{
				Type: "docker", // TODO: use containerd image store when it is integrated to Docker
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return w, nil
				},
			},
		},
		LocalMounts: map[string]fsutil.FS{
			"context":    cxtLocalMount,
			"dockerfile": dockerfileLocalMount,
		},
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": filepath.Base(config.Dockerfile),
		},
		Ref: identity.NewID(),
	}, nil
}

func (b *Builder) Close() error {
	return b.daemon.Close()
}

func loadDockerTar(r io.Reader) (string, error) {
	// no need to use moby/moby/client here
	cmd := exec.Command("docker", "load")
	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdin = r
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`Loaded image ID:\s+(\S+)`)
	m := re.FindStringSubmatch(stdoutBuffer.String())
	if len(m) < 2 {
		return "", fmt.Errorf("couldn't find loaded image ID")
	}
	imageID := m[1]
	return imageID, nil
}
