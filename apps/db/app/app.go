package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ettle/strcase"
	"github.com/protobuf-orm/protobuf-orm/graph"
	"github.com/protobuf-orm/protobuf-orm/ormpb"
	"github.com/protobuf-orm/protoc-gen-orm-dexie/internal/build"
	"google.golang.org/protobuf/compiler/protogen"
)

type App struct {
	build.App
}

func New(output string) (*App, error) {
	return &App{}, nil
}

func (a *App) Run(ctx context.Context, p *protogen.Plugin, frame *build.Frame) error {
	{
		gf := a.NewGeneratedFile(p, frame, "db.g.ts")
		a.xDb(gf, frame)
	}
	for _, def := range frame.Defs {
		if def.ServiceFile == nil {
			continue
		}

		gf := a.NewGeneratedFile(p, frame, def.Name()+".db.g.ts")
		a.xDefDb(gf, def)
	}

	return nil
}

func (a *App) xDb(f *protogen.GeneratedFile, frame *build.Frame) error {
	for _, def := range frame.Defs {
		x_name := def.Entity.Name()
		f.P(`import * as `, x_name, ` from "./`, def.Name()+".db.g.ts", `";`)
	}
	f.P(``)
	f.P(`export function schemas(){`)
	f.P(`	return {`)
	for _, def := range frame.Defs {
		x_name := def.Entity.Name()
		f.P(`		[`, x_name+".TableName", `]: `, x_name+".Schema", `,`)
	}
	f.P(`	} as const`)
	f.P(`}`)
	return nil
}

func (a *App) xDexieSchemaString(f *protogen.GeneratedFile, def *build.Def) string {
	sb := strings.Builder{}
	sb.WriteString(",")
	for k := range def.Entity.Keys() {
		if k == def.Entity.Key() {
			continue
		}

		switch k := k.(type) {
		case graph.Field:
			sb.WriteString("&" + k.Name())

		case graph.Index:
			names := []string{}
			for p := range k.Props() {
				name := ""
				switch p := p.(type) {
				case graph.Field:
					name = p.Name()
				case graph.Edge:
					name = p.Name() + "." + p.Target().Key().Name()
				}
				names = append(names, name)
			}

			sb.WriteString("[")
			sb.WriteString(strings.Join(names, "+"))
			sb.WriteString("]")

		default:
			panic(`unimplemented: key type not Field`)
		}
	}
	return sb.String()
}

func (a *App) xDefDb(f *protogen.GeneratedFile, def *build.Def) error {
	x_key := def.Entity.Key()
	x_key_t := x_key.Type()
	x_name := def.Entity.Name()
	x_schema := x_name + "Schema"

	_keyer := func(v string) string {
		return keyer(x_key_t, v)
	}

	f.P(`import { create, type MessageInitShape } from "@bufbuild/protobuf";`)
	f.P(`import { Code } from "@connectrpc/connect";`)
	f.P(`import { DbBase, type DbOf, type EntityOf, type Key, uuid } from "@protobuf-orm/runtime";`)
	f.P(``)

	f.P(`import type { `, x_name+`ServiceClient`, ` } from "./client.g"`)
	f.P(`import { type `, x_name, `, `, x_name+`Schema`, ` } from "./`, filepath.Base(def.EntityFilePath), `"`)
	f.P(`import type { `, x_name+`GetRequestSchema } from "./`, filepath.Base(def.ServiceFilepath), `"`)
	f.P(``)

	f.P(`type Desc = typeof `, x_schema)
	f.P(`export type Db = DbOf<Desc>`)
	f.P(``)

	f.P(`export const TableName = "`, def.Entity.FullName(), `";`)
	f.P(`export const Schema = "`, a.xDexieSchemaString(f, def), `" as const;`)
	f.P(``)

	f.P(`export class `, x_name+"ServiceDb", ` extends DbBase<Desc> implements Partial<`, x_name+"ServiceClient", `> {`)
	f.P(`	constructor(db: Db) {`)
	f.P(`		super(db, `, x_schema, `);`)
	f.P(`	}`)
	f.P(``)

	//#region _dehydrate
	f.P(`	_dehydrate(v: `, x_name, `): [Key, any] {`)
	f.P(`		const w = v as unknown as any`)
	switch x_key_t {
	case ormpb.Type_TYPE_UUID:
		f.P(`		const k = uuid.u8_str(`, "v."+x_key.Name(), `)`)
		f.P(`		if(k === undefined) throw this._err("invalid key", Code.InvalidArgument)`)
	default:
		f.P(`		const k = `, "v."+x_key.Name())
	}
	f.P(`		`, "w."+x_key.Name(), ` = k`)
	for k := range def.Entity.Keys() {
		if k == x_key {
			continue
		}

		switch k := k.(type) {
		case graph.Field:
			switch k.Type() {
			case ormpb.Type_TYPE_UUID:
				f.P(`		`, "w."+k.Name(), ` = uuid.u8_str(`, "v."+k.Name(), `)`)
			}

		case graph.Index:
			for p := range k.Props() {
				switch p := p.(type) {
				case graph.Field:
					switch p.Type() {
					case ormpb.Type_TYPE_UUID:
						f.P(`		`, "w."+p.Name(), ` = uuid.u8_str(`, "v."+p.Name(), `)`)
					}

				case graph.Edge:
					k_target := p.Target().Key()
					switch k_target.Type() {
					case ormpb.Type_TYPE_UUID:
						f.P(`		`, "w."+p.Name()+"."+k_target.Name(), ` = uuid.u8_str(`, "v."+p.Name()+"?."+k_target.Name(), `)`)
					}
				}
			}
		}
	}
	f.P(`		return [k, v]`)
	f.P(`	}`)
	f.P(``)
	//#endregion

	//#region _hydrate
	f.P(`	_hydrate(v: any): `, x_name, ` {`)
	for k := range def.Entity.Keys() {
		switch k := k.(type) {
		case graph.Field:
			switch k.Type() {
			case ormpb.Type_TYPE_UUID:
				f.P(`		`, "v."+k.Name(), ` = uuid.str_u8(`, "v."+k.Name(), `)`)
			}

		case graph.Index:
			for p := range k.Props() {
				switch p := p.(type) {
				case graph.Field:
					switch p.Type() {
					case ormpb.Type_TYPE_UUID:
						f.P(`		`, "v."+p.Name(), ` = uuid.str_u8(`, "v."+p.Name(), `)`)
					}

				case graph.Edge:
					y_key := p.Target().Key()
					switch y_key.Type() {
					case ormpb.Type_TYPE_UUID:
						f.P(`		`, "v."+p.Name()+"."+y_key.Name(), ` = uuid.str_u8(`, "v."+p.Name()+"?."+y_key.Name(), `)`)
					}
				}
			}
		}
	}
	f.P(`		return create(this._schema, v)`)
	f.P(`	}`)
	f.P(``)
	//#endregion

	//#region versioned
	if ver := def.Entity.GetVersionField(); ver != nil {
		v_name := ver.Name()
		va := "a." + strcase.ToCamel(v_name)
		vb := "b." + strcase.ToCamel(v_name)
		f.P(`	_versioned(): boolean {`)
		f.P(`		return true`)
		f.P(`	}`)
		f.P(``)
		f.P(`	_compare(a: EntityOf<Desc>, b: EntityOf<Desc>): number {`)
		f.P(`		if(`, va, ` === undefined) return 1`)
		f.P(`		if(`, vb, ` === undefined) return -1`)
		f.P(`		const d = `, va+".seconds", ` - `, vb+".seconds")
		f.P(`		if(d > 0) return 1`)
		f.P(`		if(d < 0) return -1`)
		f.P(`		return `, va+".nanos", ` - `, vb+".nanos")
		f.P(`	}`)
		f.P(``)
	}
	//#endregion

	//#region get
	f.P(`	get(req: MessageInitShape<typeof `, x_name+"GetRequestSchema", `>): Promise<`, x_name, `> {`)
	f.P(`		if(req.ref?.key === undefined) return Promise.reject(this._err("key undefined", Code.InvalidArgument));`)
	f.P(`		const { key } = req.ref;`)
	f.P(`		switch(key.case) {`)
	{
		f.P(`			case "`, x_key.Name(), `": {`)
		f.P(`				const k = `, _keyer(`key.value`), `;`)
		f.P(`				if(k === undefined) return Promise.reject(this._err("invalid key", Code.InvalidArgument));`)
		f.P(`				return this._query(t => t.get(k));`)
		f.P(`			}`)
		f.P(``)
	}
	for k := range def.Entity.Keys() {
		if k == x_key {
			continue
		}

		f.P(`			case "`, k.Name(), `": {`)
		switch k := k.(type) {
		case graph.Field:
			f.P(`				const k = `, keyer(k.Type(), `key.value;`))
			f.P(`				return this._query(this._table.get(k));`)

		case graph.Index:
			f.P(`				const v = key.value;`)
			f.P(`				const q = {`)
			for p := range k.Props() {
				switch p := p.(type) {
				case graph.Field:
					f.P(`					'`, p.Name(), `': `, "v."+p.Name(), `,`)
				case graph.Edge:
					y_key := p.Target().Key()
					y_path := p.Name() + "." + p.Target().Key().Name()
					y_v := "v." + p.Name() + "?.key?.value"
					switch y_key.Type() {
					case ormpb.Type_TYPE_UUID:
						y_v = fmt.Sprintf("uuid.u8_str(%s)", y_v)
					}
					f.P(`					'`, y_path, `': `, y_v, `,`)
				}
			}
			f.P(`				}`)
			f.P(`				return this._query(t => t.where(q).first());`)

		default:
			panic(`unimplemented: key type not Field`)
		}
		f.P(`			}`)
		f.P(``)
	}
	f.P(`			default:`)
	f.P("				return Promise.reject(this._err(`unknown key: ${key.case}`, Code.InvalidArgument));")
	f.P(`		}`)
	f.P(`	}`)
	f.P(`}`)
	//#endregion

	return nil
}

func keyer(t ormpb.Type, v string) string {
	switch t {
	case ormpb.Type_TYPE_UUID:
		return fmt.Sprintf("uuid.u8_str(%s)", v)
	}

	return v
}
