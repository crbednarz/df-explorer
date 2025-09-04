package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/appcontext"
)

type InstructionType string

const (
	InstructionAdd         InstructionType = "ADD"
	InstructionArg         InstructionType = "ARG"
	InstructionCmd         InstructionType = "CMD"
	InstructionCopy        InstructionType = "COPY"
	InstructionEntrypoint  InstructionType = "ENTRYPOINT"
	InstructionEnv         InstructionType = "ENV"
	InstructionExpose      InstructionType = "EXPOSE"
	InstructionFrom        InstructionType = "FROM"
	InstructionHealthcheck InstructionType = "HEALTHCHECK"
	InstructionLabel       InstructionType = "LABEL"
	InstructionMaintainer  InstructionType = "MAINTAINER"
	InstructionOnbuild     InstructionType = "ONBUILD"
	InstructionRun         InstructionType = "RUN"
	InstructionShell       InstructionType = "SHELL"
	InstructionStopsignal  InstructionType = "STOPSIGNAL"
	InstructionUser        InstructionType = "USER"
	InstructionVolume      InstructionType = "VOLUME"
	InstructionWorkdir     InstructionType = "WORKDIR"
)

type Dockerfile struct {
	buildConfig BuildConfig
	definition  *llb.Definition
}

func NewDockerfile(buildContext string, dockerfile string) (*Dockerfile, error) {
	df := &Dockerfile{
		buildConfig: BuildConfig{
			BuildContext: buildContext,
			Dockerfile:   dockerfile,
		},
	}
	err := df.reload()
	return df, err
}

func (df *Dockerfile) Build(ctx context.Context, builder *Builder) (string, error) {
	imageID, err := builder.Build(ctx, df.buildConfig)
	if err != nil {
		return imageID, fmt.Errorf("unable to build image: %w", err)
	}

	return imageID, nil
}

func (df *Dockerfile) Append(line string) error {
	lineToAppend := append([]byte(line), byte('\n'), byte('\n'))
	f, err := os.OpenFile(df.buildConfig.Dockerfile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open dockerfile: %w", err)
	}

	if _, err := f.Write(lineToAppend); err != nil {
		return fmt.Errorf("unable to write dockerfile: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("unable to close dockerfile: %w", err)
	}
	return df.reload()
}

func (df *Dockerfile) reload() error {
	rawFile, err := os.ReadFile(df.buildConfig.Dockerfile)
	caps := pb.Caps.CapSet(pb.Caps.All())

	fileName := filepath.Base(df.buildConfig.Dockerfile)
	sourceMap := llb.NewSourceMap(nil, fileName, "Dockerfile", rawFile)

	state, _, _, _, err := dockerfile2llb.Dockerfile2LLB(appcontext.Context(), rawFile, dockerfile2llb.ConvertOpt{
		MetaResolver: imagemetaresolver.Default(),
		LLBCaps:      &caps,
		Config: dockerui.Config{
			Target: "",
		},
		SourceMap: sourceMap,
	})
	if err != nil {
		return err
	}

	definition, err := state.Marshal(context.TODO())
	if err != nil {
		return err
	}
	df.definition = definition
	return nil
}
