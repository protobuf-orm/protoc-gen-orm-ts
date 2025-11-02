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
	Defs []*Def
}

type Def struct {
	Entity  graph.Entity
	Service *protogen.Service

	EntityFile  *protogen.File
	ServiceFile *protogen.File

	EntityFilePath  string
	ServiceFilepath string
}

func (d *Def) Name() string {
	return strings.ToLower(d.Entity.Name())
}

func (d *Def) Path(pkg string) string {
	return filepath.Join(pkg, d.Name())
}

func ParseFrame(p *protogen.Plugin, g *graph.Graph) (*Frame, error) {
	frame := &Frame{G: g}
	defs := map[string]*Def{}
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}
		for _, m := range f.Messages {
			entity, ok := g.Entities[m.Desc.FullName()]
			if !ok {
				continue
			}

			def := &Def{
				Entity:     entity,
				EntityFile: f,
			}
			defs[entity.Name()] = def
			frame.Defs = append(frame.Defs, def)

			// path/to/package/xxx.proto
			p := f.Desc.Path()
			name_f, _ := strings.CutSuffix(p, filepath.Ext(p))
			def.EntityFilePath = name_f + "_pb.ts"
		}
	}

	var f_ *protogen.File
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}
		if len(f.Services) == 0 {
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

			def, ok := defs[name]
			if !ok {
				panic(fmt.Sprintf("service name not found in Entities: %s", name))
			}

			// path/to/package/xxx_svc.g.proto
			p := f.Desc.Path()
			name_f, _ := strings.CutSuffix(p, filepath.Ext(p))

			def.Service = s
			def.ServiceFile = f
			def.ServiceFilepath = name_f + "_pb.ts"
		}
	}
	if f_ == nil {
		return nil, io.EOF
	}

	// path/to/package
	frame.Path = strings.ReplaceAll(string(f_.Desc.Package()), ".", "/")
	return frame, nil
}
