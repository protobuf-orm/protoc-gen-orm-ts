package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/protobuf-orm/protobuf-orm/graph"
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
	gf.P(`import type { QueryDescOf } from "@protobuf-orm/runtime";`)
	gf.P(``)

	for _, info := range frame.Services {
		filename := filepath.Base(info.Path)
		svc_name := info.Def.Desc.Name()
		gf.P(`import { `, svc_name, ` } from "./`, filename, `";`)
	}
	gf.P(``)
	for _, info := range frame.Services {
		svc_name := info.Def.Desc.Name()
		gf.P(`export type `, svc_name, `Client = C<typeof `, svc_name, `>`)
	}
	gf.P(``)
	gf.P(`export interface ServiceClient {`)
	for _, info := range frame.Services {
		svc_name := info.Def.Desc.Name()
		name, _ := strings.CutSuffix(string(svc_name), "Service")
		gf.P(`	readonly `, camel(name), `: `, svc_name, `Client;`)
	}
	gf.P(`}`)
	gf.P(``)
	for _, info := range frame.Services {
		svc_name := info.Def.Desc.Name()
		name, _ := strings.CutSuffix(string(svc_name), "Service")
		gf.P(`export const `, camel(name), `= `, svc_name, `.method;`)
	}
	gf.P(``)
	gf.P(`export const queries = {`)
	for _, info := range frame.Services {
		if info.Entity == nil {
			continue
		}

		svc_name := info.Def.Desc.Name()
		name, _ := strings.CutSuffix(string(svc_name), "Service")
		gf.P(`	["`, string(info.Def.Desc.ParentFile().Package())+"."+string(svc_name), `"]: {`)
		gf.P(`		pick: v => ({key:{`, genPick(info.Entity.Def.Key().Name()), `}}),`)
		gf.P(`		refs: v => [`)
		for k := range info.Entity.Def.Keys() {
			switch k := k.(type) {
			case graph.Field:
				gf.P(`			{key:{`, genPick(k.Name()), `}},`)
			case graph.Index:
				gf.P(`			{key:{`)
				gf.P(`				case: "`, k.Name(), `",`)
				gf.P(`				value: {`)
				for p := range k.Props() {
					v_path := "v." + p.Name()
					switch p := p.(type) {
					case graph.Field:
						gf.P(`					`, p.Name(), `: `, v_path, `,`)
					case graph.Edge:
						target_name := p.Target().Key().Name()
						gf.P(`					`, p.Name(), `: `, v_path, ` && {key:{`, genPickWithIdent(target_name, v_path), `}},`)
					}
				}
				gf.P(`				}`)
				gf.P(`			}}`)
			}
		}
		gf.P(`		],`)
		gf.P(`		rpc: {`)
		for _, method := range info.Def.Methods {
			method_name := string(method.Desc.Name())
			gf.P(`			`, camel(method_name), `: {`)
			gf.P(`				desc: `, camel(name)+"."+camel(method_name), `,`)
			gf.P(`				extract: v => `, genExtractor(info.Entity, method))
			gf.P(`			},`)
		}
		gf.P(`		}`)
		gf.P(`	} satisfies QueryDescOf<typeof `, svc_name, `>,`)
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

func genPick(k string) string {
	return genPickWithIdent(k, "v")
}

func genPickWithIdent(k string, v string) string {
	return fmt.Sprintf("case: %q, value: %s.%s", k, v, k)
}

func genExtractor(entity *build.EntityInfo, m *protogen.Method) string {
	entity_desc := entity.Def.Descriptor()
	if m.Output.Desc == entity_desc {
		return "v"
	}
	for i := 0; i < m.Output.Desc.Fields().Len(); i++ {
		field := m.Output.Desc.Fields().Get(i)
		if field.Message() != entity_desc {
			continue
		}

		return "v." + string(field.Name())
	}
	return "undefined"
}
