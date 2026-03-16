package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	buildkit "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/opencontainers/go-digest"
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
	buildContext string
	dockerfile   string
	definition   *llb.Definition
	state        *llb.State
	source       *Source
	imageID      string
	vertexMap    map[string]llb.Vertex
}

type Source struct {
	Sections []SourceSection
}

type SourceSection struct {
	Text       string
	VertexHash string
	Metadata   *llb.OpMetadata
}

func NewDockerfile(buildContext string, dockerfile string) (*Dockerfile, error) {
	df := &Dockerfile{
		buildContext: buildContext,
		dockerfile:   dockerfile,
	}
	err := df.reload()
	return df, err
}

func (df *Dockerfile) ImageID() string {
	return df.imageID
}

func (df *Dockerfile) Source() *Source {
	return df.source
}

func (df *Dockerfile) FileName() string {
	return filepath.Base(df.dockerfile)
}

func (df *Dockerfile) Path() string {
	return df.dockerfile
}

func (df *Dockerfile) Dir() string {
	return filepath.Dir(df.dockerfile)
}

func (df *Dockerfile) Build(ctx context.Context, builder *Builder, progress chan *buildkit.SolveStatus) (string, error) {
	return df.build(ctx, builder, df.definition, progress)
}

func (df *Dockerfile) BuildToVertex(ctx context.Context, builder *Builder, vertex string, progress chan *buildkit.SolveStatus) (string, error) {
	subDef, err := df.getSubDefinition(vertex)
	if err != nil {
		return "", err
	}
	return df.build(ctx, builder, subDef, progress)
}

func (df *Dockerfile) build(ctx context.Context, builder *Builder, definition *llb.Definition, progress chan *buildkit.SolveStatus) (string, error) {
	imageID, err := builder.Build(
		ctx,
		WithDockerfile(df.dockerfile, df.buildContext),
		WithDefinition(definition),
		WithProgressChannel(progress),
	)
	if err != nil {
		return imageID, fmt.Errorf("unable to build image: %w", err)
	}

	df.imageID = imageID
	return imageID, nil
}

func (df *Dockerfile) getSubDefinition(vertex string) (*llb.Definition, error) {
	v, ok := df.vertexMap[vertex]
	if !ok {
		return nil, fmt.Errorf("failed to find vertex: %v", vertex)
	}
	state := llb.NewState(v.Output())
	return state.Marshal(context.TODO())
}

func (df *Dockerfile) createVertexMap() map[string]llb.Vertex {
	vertexMap := make(map[string]llb.Vertex, 0)
	populateVertexMap(context.TODO(), vertexMap, df.state.Output(), df.definition.Constraints)
	return vertexMap
}

func populateVertexMap(ctx context.Context, vertexMap map[string]llb.Vertex, output llb.Output, constraints *llb.Constraints) {
	v := output.Vertex(ctx, constraints)
	dgst, _, _, _, err := v.Marshal(ctx, constraints)
	if err != nil {
		return
	}
	vertexMap[dgst.String()] = v
	for _, input := range v.Inputs() {
		populateVertexMap(ctx, vertexMap, input, constraints)
	}
}

func (df *Dockerfile) Append(line string) error {
	lineToAppend := append([]byte(line), byte('\n'), byte('\n'))
	f, err := os.OpenFile(df.dockerfile, os.O_APPEND|os.O_WRONLY, 0o644)
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
	rawFile, err := os.ReadFile(df.dockerfile)
	if err != nil {
		return fmt.Errorf("unable to read dockerfile: %w", err)
	}
	caps := pb.Caps.CapSet(pb.Caps.All())

	fileName := filepath.Base(df.dockerfile)
	sourceMap := llb.NewSourceMap(nil, fileName, "Dockerfile", rawFile)

	state, _, _, _, err := dockerfile2llb.Dockerfile2LLB(appcontext.Context(), rawFile, dockerfile2llb.ConvertOpt{
		MetaResolver: imagemetaresolver.Default(),
		LLBCaps:      &caps,
		Config: dockerui.Config{
			Target: "",
		},
		SourceMap: sourceMap,
	})
	df.state = state
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

	df.imageID = ""
	df.vertexMap = df.createVertexMap()
	return nil
}

func (df *Dockerfile) rebuildSourceMap() error {
	rawDockerfile, err := os.ReadFile(df.dockerfile)
	if err != nil {
		return fmt.Errorf("unable to read dockerfile during source map rebuild: %w", err)
	}

	lines, err := parseDockerfileLines(df.definition, string(rawDockerfile))
	if err != nil {
		return fmt.Errorf("unable to parse dockerfile lines: %w", err)
	}

	var sections []SourceSection
	lastMeta := lines[0].Metadata
	lastVertex := lines[0].VertexHash
	sectionStart := 0

	for i := 1; i < len(lines); i++ {
		meta := lines[i].Metadata
		if meta != lastMeta {
			sections = append(sections, SourceSection{
				Text:       joinLines(lines[sectionStart:i]),
				Metadata:   lastMeta,
				VertexHash: string(lastVertex),
			})
			lastMeta = meta
			lastVertex = lines[i].VertexHash
			sectionStart = i
		}
	}
	sections = append(sections, SourceSection{
		Text:       joinLines(lines[sectionStart:]),
		Metadata:   lastMeta,
		VertexHash: string(lastVertex),
	})

	df.source = &Source{
		Sections: sections,
	}
	return nil
}

type lineMetadata struct {
	Text       string
	Metadata   *llb.OpMetadata
	VertexHash digest.Digest
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
					lines[i-1].VertexHash = hash
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
