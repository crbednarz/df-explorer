package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	source      *Source
}

type Source struct {
	Chunks []SourceChunk
}

type SourceChunk struct {
	Text     string
	Metadata *llb.OpMetadata
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

	err = df.rebuildSourceMap()
	if err != nil {
		return err
	}

	return nil
}

func (df *Dockerfile) rebuildSourceMap() error {
	rawDockerfile, err := os.ReadFile(df.buildConfig.Dockerfile)
	if err != nil {
		return fmt.Errorf("unable to read dockerfile during source map rebuild: %w", err)
	}

	lines, err := parseDockerfileLines(df.definition, string(rawDockerfile))
	if err != nil {
		return fmt.Errorf("unable to parse dockerfile lines: %w", err)
	}

	var chunks []SourceChunk
	lastMeta := lines[0].Metadata
	chunkStart := 0

	for i := 1; i < len(lines); i++ {
		meta := lines[i].Metadata
		if meta != lastMeta {
			chunks = append(chunks, SourceChunk{
				Text:     joinLines(lines[chunkStart:i]),
				Metadata: lastMeta,
			})
			lastMeta = meta
			chunkStart = i
		}
	}
	chunks = append(chunks, SourceChunk{
		Text:     joinLines(lines[chunkStart:]),
		Metadata: lastMeta,
	})

	df.source = &Source{
		Chunks: chunks,
	}
	return nil
}

func (df *Dockerfile) Source() *Source {
	return df.source
}

type lineMetadata struct {
	Text     string
	Metadata *llb.OpMetadata
}

func parseDockerfileLines(definition *llb.Definition, rawDockerfile string) ([]lineMetadata, error) {
	sourceLines := strings.Split(string(rawDockerfile), "\n")
	lines := make([]lineMetadata, len(sourceLines))

	for i, line := range sourceLines {
		lines[i] = lineMetadata{
			Text: line,
		}
	}

	for hash, meta := range definition.Metadata {
		hashStr := string(hash)
		sourceLocations, ok := definition.Source.Locations[hashStr]
		if !ok {
			continue
		}

		for _, loc := range sourceLocations.Locations {
			for _, locRange := range loc.Ranges {
				for i := locRange.Start.Line; i <= locRange.End.Line; i++ {
					lines[i-1].Metadata = &meta
				}
			}
		}
	}

	writeIndex := 0
	for i := 0; i < len(lines); i++ {
		// TODO: Seems wasteful to allocate here
		if strings.TrimSpace(lines[i].Text) == "" {
			continue
		}
		if i != writeIndex {
			lines[writeIndex] = lines[i]
		}
		writeIndex++
	}
	return lines[:writeIndex], nil
}

func joinLines(lines []lineMetadata) string {
	if len(lines) == 0 {
		return ""
	} else if len(lines) == 1 {
		return lines[0].Text
	}

	joinedSize := len(lines) - 1
	for _, line := range lines {
		joinedSize += len(line.Text)
	}

	var b strings.Builder
	b.Grow(joinedSize)
	b.WriteString(lines[0].Text)
	for _, line := range lines[1:] {
		b.WriteString("\n")
		b.WriteString(line.Text)
	}
	return b.String()
}
