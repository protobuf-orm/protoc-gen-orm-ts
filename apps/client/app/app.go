package app

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/protobuf-orm/protoc-gen-orm-ts/internal/build"
	"google.golang.org/protobuf/compiler/protogen"
)

type App struct {
	build.App
}

func New(output string) (*App, error) {
	return &App{}, nil
}

func (a *App) Run(ctx context.Context, p *protogen.Plugin, frame *build.Frame) error {
	gf := a.NewGeneratedFile(p, frame, "client.g.ts")
	gf.P(`import type { Client as C } from "@connectrpc/connect";`)
	gf.P(``)

	for _, def := range frame.Defs {
		filename := filepath.Base(def.ServiceFilepath)
		svc_name := def.Service.Desc.Name()
		gf.P(`import type { `, svc_name, ` } from "./`, filename, `";`)
	}
	gf.P(``)
	for _, def := range frame.Defs {
		svc_name := def.Service.Desc.Name()
		gf.P(`export type `, svc_name, `Client = C<typeof `, svc_name, `>`)
	}
	gf.P(``)
	gf.P(`export interface ServiceClient {`)
	for _, def := range frame.Defs {
		svc_name := def.Service.Desc.Name()
		name, _ := strings.CutSuffix(string(svc_name), "Service")
		gf.P(`	readonly `, camel(name), `: `, svc_name, `Client;`)
	}
	gf.P(`}`)
	gf.P(``)
	return nil
}

// Very naive implementation.
// It works though since the name of the entity is PascalCase.
func camel(v string) string {
	return strings.ToLower(v[:1]) + v[1:]
}
