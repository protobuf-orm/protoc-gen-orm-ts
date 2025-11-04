import { clone, type DescMessage, type MessageShape } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import type Dexie from "dexie";
import type { Table } from "dexie";

import * as unsafe from "./unsafe";

export type Key = string | number;

type V_<Desc extends DescMessage> = MessageShape<Desc>;

// Entity type is a type of dehydrated value which does not hold
// known constant values like $typename and have indexable types for the key fields.
// I couldn't make the exact dehydrated type with an indexable key type at this moment.
type E_<Desc extends DescMessage> = Omit<V_<Desc>, "$typeName" | "$unknown">;

type T_<Desc extends DescMessage> = Table<E_<Desc>, Key>;
type D_<Desc extends DescMessage> = Dexie &
	Record<V_<Desc>["$typeName"], T_<Desc>>;

export type ValueOf<Desc extends DescMessage> = V_<Desc>;
export type EntityOf<Desc extends DescMessage> = E_<Desc>;
export type DbOf<Desc extends DescMessage> = D_<Desc>;

export class TableBase<Desc extends DescMessage = DescMessage> {
	protected readonly _db: D_<Desc>;
	readonly _schema: Desc;
	constructor(db: D_<Desc>, schema: Desc) {
		this._db = db;
		this._schema = schema;
	}

	get _typeName(): V_<Desc>["$typeName"] {
		return this._schema.typeName;
	}
	get _table(): T_<Desc> {
		return this._db[this._typeName];
	}

	protected _err(msg: string, code: Code) {
		return new ConnectError(`${this._typeName}: ${msg}`, code);
	}

	// TODO: I think de/hydrate function can be constructed from the schema at the runtime.
	protected _dehydrate(v: V_<Desc>): [Key, any] {
		throw new Error("virtual function not implemented");
	}
	protected _hydrate(v: any): V_<Desc> {
		throw new Error("virtual function not implemented");
	}
	protected _versioned(): boolean {
		return false;
	}

	// It does not compares the value or entity but their version
	// so it assumes two values have same ID.
	_compare(a: V_<Desc> | E_<Desc>, b: V_<Desc> | E_<Desc>): number {
		return -1;
	}

	private _makeDehydrated(v: V_<Desc>): [Key, any] {
		v = clone(this._schema, v);
		const res = this._dehydrate(v);
		unsafe.rm(res[1], "$typeName");
		unsafe.rm(res[1], "$unknown");
		return res;
	}

	_query(q: (t: T_<Desc>) => Promise<E_<Desc> | undefined>): Promise<V_<Desc>> {
		return q(this._table).then((v) => {
			if (v === undefined) throw this._err("not found", Code.NotFound);
			return this._hydrate(v);
		});
	}
	async _insert(v: V_<Desc>) {
		const [k, data] = this._makeDehydrated(v);

		return this._table.add(data, k);
	}
	async _reconcile(v: V_<Desc>): Promise<boolean> {
		const [k, data] = this._makeDehydrated(v);

		if (!this._versioned()) {
			await this._table.put(data, k);
			return true;
		}

		await this._db.transaction("rw", this._table, async (_tx) => {
			const u = await this._table.get(k);
			if (u === undefined) {
				return this._table.put(data, k);
			}
			if (this._compare(u, v) >= 0) {
				return false;
			}
			return this._table.put(data, k);
		});
		return true;
	}
}
