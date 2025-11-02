import { create } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, test } from "vitest";
import "fake-indexeddb/auto";
import { timestampFromDate, timestampNow } from "@bufbuild/protobuf/wkt";
import Dexie from "dexie";
import { IDBFactory } from "fake-indexeddb";
import { TenantSchema } from "@/gen/apptest/tenant_pb";
import * as UserDb from "@/gen/apptest/user.db.g";
import { UserSchema } from "@/gen/apptest/user_pb";
import { uuid } from ".";

describe("db", () => {
	let db: UserDb.Db;
	const tenant = create(TenantSchema, {
		id: uuid.str_u8("b45df899-1672-44fe-8d07-7ca1f9dabc62"),
		alias: "hday",
		name: "Holiday Robotics",
	});
	const v = create(UserSchema, {
		id: uuid.str_u8("fff7ed22-8ad5-4e7b-b976-9d1e8ea74021"),
		tenant,
		alias: "hal",
		name: "HAL",
		dateUpdated: timestampFromDate(new Date(2001, 0, 1)),
		dateCreated: timestampFromDate(new Date(1992, 0, 12)),
	});
	const ref = {
		key: {
			case: "id",
			value: v.id,
		},
	} as const;

	beforeEach(async () => {
		const indexedDB = new IDBFactory();
		db = new Dexie("test", { indexedDB }) as UserDb.Db;
		db.version(1).stores({
			[UserDb.TableName]: UserDb.Schema,
		});
	});

	test("get by primary key", async () => {
		const userDb = new UserDb.UserServiceDb(db);
		await userDb._insert(v);

		const result = await userDb.get({
			ref: {
				key: {
					case: "id",
					value: v.id,
				},
			},
		});
		expect(result).toEqual(v);
	});
	test("get by unique index", async () => {
		const userDb = new UserDb.UserServiceDb(db);
		await userDb._insert(v);

		const result = await userDb.get({
			ref: {
				key: {
					case: "alias",
					value: {
						tenant: {
							key: {
								case: "id",
								value: tenant.id,
							},
						},
						alias: v.alias,
					},
				},
			},
		});
		expect(result).toEqual(v);
	});
	test("reconcile with new value", async () => {
		const userDb = new UserDb.UserServiceDb(db);
		await userDb._reconcile(v);

		const result = await userDb.get({ ref });
		expect(result).toEqual(v);
	});
	test("reconcile with newer value", async () => {
		const userDb = new UserDb.UserServiceDb(db);
		await userDb._insert(v);

		const w = create(UserSchema, { ...v, dateUpdated: timestampNow() });
		await userDb._reconcile(w);

		const result = await userDb.get({ ref });
		expect(result).toEqual(w);
	});
	test("reconcile with older value", async () => {
		const userDb = new UserDb.UserServiceDb(db);

		const w = create(UserSchema, { ...v, dateUpdated: timestampNow() });
		await userDb._insert(w);
		await userDb._reconcile(v);

		const result = await userDb.get({ ref });
		expect(result).toEqual(w);
	});
});
