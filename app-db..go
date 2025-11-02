package main

import (
	"context"
	"fmt"

	"github.com/protobuf-orm/protoc-gen-orm-ts/apps/db/app"
	"github.com/protobuf-orm/protoc-gen-orm-ts/internal/build"
	"google.golang.org/protobuf/compiler/protogen"
)

type DbOpts struct{}

func (h *DbOpts) Run(ctx context.Context, p *protogen.Plugin, output string, frame *build.Frame) error {
	app, err := app.New(output)
	if err != nil {
		return fmt.Errorf("initialize client app: %w", err)
	}
	if err := app.Run(ctx, p, frame); err != nil {
		return fmt.Errorf("run schema app: %w", err)
	}

	return nil
}
