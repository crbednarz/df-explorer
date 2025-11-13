package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	"github.com/crbednarz/df-explorer/pkg/util"
	dclient "github.com/docker/docker/client"
	buildkit "github.com/moby/buildkit/client"
	_ "github.com/moby/buildkit/client/connhelper/dockercontainer"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/identity"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/sync/errgroup"
)

type Builder struct {
	client *buildkit.Client
	daemon Container
}

type BuildProgressCallback func()

type BuildConfig struct {
	BuildContext    string
	Dockerfile      string
	Definition      *llb.Definition
	ProgressChannel chan *buildkit.SolveStatus
}

type BuildOption func(*BuildConfig)

func WithDockerfile(dockerfile string, contextPath string) BuildOption {
	return func(config *BuildConfig) {
		config.Dockerfile = dockerfile
		config.BuildContext = contextPath
	}
}

func WithDefinition(def *llb.Definition) BuildOption {
	return func(config *BuildConfig) {
		config.Definition = def
	}
}

func WithProgressChannel(channel chan *buildkit.SolveStatus) BuildOption {
	return func(config *BuildConfig) {
		config.ProgressChannel = channel
	}
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
		WithRemoveOnClean(false),
		WithReuse(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create buildkit daemon container: %w", err)
	}

	c, err := buildkit.New(ctx, "docker-container://df-buildkitd")
	if err != nil {
		return nil, err
	}

	return &Builder{
		client: c,
		daemon: daemon,
	}, nil
}

func (b *Builder) Build(ctx context.Context, buildOptions ...BuildOption) (string, error) {
	config := BuildConfig{}
	for _, opt := range buildOptions {
		opt(&config)
	}
	pipeR, pipeW := io.Pipe()
	solveOpt, err := newSolveOpt(pipeW, config)
	if err != nil {
		return "", err
	}
	var imageID string
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		_, err := b.client.Solve(ctx, config.Definition, *solveOpt, config.ProgressChannel)
		if err != nil {
			pipeW.CloseWithError(err)
		}
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

func newSolveOpt(w io.WriteCloser, config BuildConfig) (*buildkit.SolveOpt, error) {
	cxtLocalMount, err := fsutil.NewFS(config.BuildContext)
	if err != nil {
		return nil, fmt.Errorf("invalid build context dir (%s): %w", config.BuildContext, err)
	}

	dockerfileLocalMount, err := fsutil.NewFS(filepath.Dir(config.Dockerfile))
	if err != nil {
		return nil, fmt.Errorf("invalid dockerfile dir (%s): %w", config.Dockerfile, err)
	}

	frontend := "dockerfile.v0"
	frontendAttrs := map[string]string{
		"filename": filepath.Base(config.Dockerfile),
	}
	if config.Definition != nil {
		frontend = ""
		frontendAttrs = map[string]string{}
	}
	return &buildkit.SolveOpt{
		Exports: []buildkit.ExportEntry{
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
		Frontend:      frontend,
		FrontendAttrs: frontendAttrs,
		Ref:           identity.NewID(),
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
