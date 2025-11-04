package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/protobuf-orm/protobuf-orm/graph"
	"github.com/protobuf-orm/protoc-gen-orm-ts/internal/build"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Handler struct {
	Client ClientOpts
	Db     DbOpts
}

func (h *Handler) Run(p *protogen.Plugin) error {
	p.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
	p.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_MAX
	p.SupportedFeatures = uint64(0 |
		pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL |
		pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS,
	)

	ctx := context.Background()
	// TODO: set logger

	g := graph.NewGraph()
	for _, f := range p.Files {
		if err := graph.Parse(ctx, g, f.Desc); err != nil {
			return fmt.Errorf("parse entity at %s: %w", *f.Proto.Name, err)
		}
	}

	frame, err := build.ParseFrame(p, g)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("parse build frame: %w", err)
	}

	h.Client.Run(ctx, p, "", frame)
	h.Db.Run(ctx, p, "", frame)
	return nil
}
