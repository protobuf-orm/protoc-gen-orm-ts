package build

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/protobuf-orm/protobuf-orm/graph"
	"google.golang.org/protobuf/compiler/protogen"
)

type Frame struct {
	G *graph.Graph

	Path string

	Entities []*EntityInfo
	Services []*ServiceInfo
}

type EntityInfo struct {
	Def  graph.Entity
	File *protogen.File
	Path string

	Service *ServiceInfo
}

func (i *EntityInfo) Name() string {
	return strings.ToLower(i.Def.Name())
}

type ServiceInfo struct {
	Def  *protogen.Service
	File *protogen.File
	Path string

	Entity *EntityInfo
}

func ParseFrame(p *protogen.Plugin, g *graph.Graph) (*Frame, error) {
	frame := &Frame{G: g}
	entities := map[string]*EntityInfo{}
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}
		for _, m := range f.Messages {
			entity, ok := g.Entities[m.Desc.FullName()]
			if !ok {
				continue
			}

			// path/to/package/xxx.proto
			p := f.Desc.Path()
			name_f, _ := strings.CutSuffix(p, filepath.Ext(p))

			info := &EntityInfo{
				Def:  entity,
				File: f,
				Path: name_f + "_pb.ts",
			}
			entities[entity.Name()] = info
			frame.Entities = append(frame.Entities, info)
		}
	}

	var f_ *protogen.File
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}

		f_ = f
		for _, s := range f.Services {
			// XxxService
			name := string(s.Desc.Name())
			name, ok := strings.CutSuffix(name, "Service")
			if !ok {
				panic(fmt.Sprintf("invalid name: %s", name))
			}

			// path/to/package/xxx_svc.g.proto
			p := f.Desc.Path()
			name_f, _ := strings.CutSuffix(p, filepath.Ext(p))

			info := &ServiceInfo{
				Def:  s,
				File: f,
				Path: name_f + "_pb.ts",
			}
			frame.Services = append(frame.Services, info)

			entity_info, ok := entities[name]
			if ok {
				entity_info.Service = info
				info.Entity = entity_info
			}
		}
	}
	if f_ == nil {
		return nil, io.EOF
	}

	// path/to/package
	frame.Path = strings.ReplaceAll(string(f_.Desc.Package()), ".", "/")
	return frame, nil
}
