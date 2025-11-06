package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ettle/strcase"
	"github.com/protobuf-orm/protobuf-orm/graph"
	"github.com/protobuf-orm/protobuf-orm/ormpb"
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
	{
		gf := a.NewGeneratedFile(p, frame, "db.g.ts")
		a.xDb(gf, frame)
	}
	for _, info := range frame.Entities {
		if info.Service == nil {
			continue
		}

		gf := a.NewGeneratedFile(p, frame, info.Name()+".db.g.ts")
		a.xDefDb(gf, info)
	}

	return nil
}

func (a *App) xDb(f *protogen.GeneratedFile, frame *build.Frame) error {
	infos := []*build.EntityInfo{}
	for _, def := range frame.Entities {
		if def.Service == nil {
			continue
		}
		infos = append(infos, def)
	}
	if len(infos) == 0 {
		f.Skip()
		return nil
	}

	for _, info := range infos {
		x_name := info.Def.Name()
		f.P(`import * as `, x_name, ` from "./`, info.Name()+".db.g.ts", `";`)
	}
	f.P(``)
	f.P(`export type Db = `)
	for _, info := range infos {
		x_name := info.Def.Name()
		f.P(`	& `, x_name+".Db")
	}
	f.P(``)
	f.P(`export interface DbClient {`)
	for _, info := range infos {
		x_name := info.Def.Name()
		f.P(`	readonly `, info.Name(), `: `, x_name+".TableService")
	}
	f.P(`}`)
	f.P(``)

	f.P(`export const DbService = {`)
	for _, info := range infos {
		x_name := info.Def.Name()
		f.P(`	[`, x_name+".TableName", `]: `, x_name+".TableService", `,`)
	}
	f.P(`}`)
	f.P(``)

	f.P(`export function schemas(){`)
	f.P(`	return {`)
	for _, info := range infos {
		x_name := info.Def.Name()
		f.P(`		[`, x_name+".TableName", `]: `, x_name+".Schema", `,`)
	}
	f.P(`	} as const`)
	f.P(`}`)
	return nil
}

func (a *App) xDexieSchemaString(f *protogen.GeneratedFile, info *build.EntityInfo) string {
	sb := strings.Builder{}
	for k := range info.Def.Keys() {
		if k == info.Def.Key() {
			continue
		}

		sb.WriteString(",")
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

func (a *App) xDefDb(f *protogen.GeneratedFile, info *build.EntityInfo) error {
	x_key := info.Def.Key()
	x_key_t := x_key.Type()
	x_name := info.Def.Name()
	x_schema := x_name + "Schema"

	_keyer := func(v string) string {
		return keyer(x_key_t, v)
	}

	f.P(`import { create, type MessageInitShape } from "@bufbuild/protobuf";`)
	f.P(`import { Code } from "@connectrpc/connect";`)
	f.P(`import type { DbOf, EntityOf, Key, ValueOf } from "@protobuf-orm/runtime";`)
	f.P(`import { TableBase, uuid } from "@protobuf-orm/runtime";`)
	f.P(``)

	f.P(`import type { `, x_name+`ServiceClient`, ` } from "./client.g"`)
	f.P(`import { type `, x_name, `, `, x_name+`Schema`, ` } from "./`, filepath.Base(info.Path), `"`)
	f.P(`import type { `, x_name+`GetRequestSchema } from "./`, filepath.Base(info.Service.Path), `"`)
	f.P(``)

	f.P(`type Desc = typeof `, x_schema)
	f.P(`export type Db = DbOf<Desc>`)
	f.P(``)

	f.P(`export const TableName = "`, info.Def.FullName(), `";`)
	f.P(`export const Schema = "`, a.xDexieSchemaString(f, info), `" as const;`)
	f.P(``)

	f.P(`export class TableService extends TableBase<Desc> implements Partial<`, x_name+"ServiceClient", `> {`)
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
	for k := range info.Def.Keys() {
		if k == x_key {
			continue
		}

		name := strcase.ToCamel(k.Name())
		switch k := k.(type) {
		case graph.Field:
			switch k.Type() {
			case ormpb.Type_TYPE_UUID:
				f.P(`		`, "w."+k.Name(), ` = uuid.u8_str(`, "v."+name, `)`)
			}

		case graph.Index:
			for p := range k.Props() {
				name := strcase.ToCamel(p.Name())
				switch p := p.(type) {
				case graph.Field:
					switch p.Type() {
					case ormpb.Type_TYPE_UUID:
						f.P(`		`, "w."+p.Name(), ` = uuid.u8_str(`, "v."+name, `)`)
					}

				case graph.Edge:
					k_target := p.Target().Key()
					switch k_target.Type() {
					case ormpb.Type_TYPE_UUID:
						f.P(`		`, "w."+name+"."+k_target.Name(), ` = uuid.u8_str(`, "v."+name+"?."+k_target.Name(), `)`)
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
	for k := range info.Def.Keys() {
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
	if ver := info.Def.GetVersionField(); ver != nil {
		v_name := ver.Name()
		va := "a." + strcase.ToCamel(v_name)
		vb := "b." + strcase.ToCamel(v_name)
		f.P(`	_versioned(): boolean {`)
		f.P(`		return true`)
		f.P(`	}`)
		f.P(``)
		f.P(`	_compare(a: ValueOf<Desc> | EntityOf<Desc>, b: ValueOf<Desc> | EntityOf<Desc>): number {`)
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
	for k := range info.Def.Keys() {
		if k == x_key {
			continue
		}

		name := strcase.ToCamel(k.Name())
		f.P(`			case "`, name, `": {`)
		switch k := k.(type) {
		case graph.Field:
			f.P(`				const k = `, keyer(k.Type(), `key.value`))
			f.P(`				const q = { `, name, `: k }`)
			f.P(`				return this._query(t => t.where(q).first());`)

		case graph.Index:
			f.P(`				const v = key.value;`)
			for p := range k.Props() {
				name := strcase.ToCamel(p.Name())
				switch p := p.(type) {
				case graph.Edge:
					f.P(`				if(v.`, name+"?.key?.case", ` !== "`, p.Target().Key().Name(), `"){`)
					f.P(`					return Promise.reject(this._err("composite query with non-keyed field not supported", Code.Unimplemented))`)
					f.P(`				}`)
				}
			}

			f.P(`				const q = {`)
			for p := range k.Props() {
				name := strcase.ToCamel(p.Name())
				switch p := p.(type) {
				case graph.Field:
					f.P(`					'`, name, `': `, "v."+name, `,`)
				case graph.Edge:
					y_key := p.Target().Key()
					y_path := name + "." + p.Target().Key().Name()
					y_v := "v." + name + "?.key?.value"
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
