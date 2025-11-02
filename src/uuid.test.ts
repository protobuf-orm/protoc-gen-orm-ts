import { assert, describe, expect, it } from "vitest";
import { str_u8, u8_str } from "./uuid";

describe("uuid conversions", () => {
	it("roundtrip known uuid", () => {
		const uuid = "550e8400-e29b-41d4-a716-446655440000";

		const u = str_u8(uuid);
		assert(u !== undefined);
		expect(Array.from(u)).toEqual([
			0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66,
			0x55, 0x44, 0x00, 0x00,
		]);

		const s = u8_str(u);
		expect(s).toBe(uuid);
	});
	it("accepts different notations", () => {
		const uuidUpper = "550E8400-E29B-41D4-A716-446655440000";

		expect(str_u8(uuidUpper)).toBeDefined();
		expect(str_u8(`{${uuidUpper}}`)).toBeDefined();
		expect(str_u8(`urn:uuid:${uuidUpper}`)).toBeDefined();
	});
	it("invalid inputs return undefined", () => {
		expect(u8_str(undefined)).toBeUndefined();
		expect(u8_str(new Uint8Array([1, 2, 3]))).toBeUndefined();
		expect(str_u8("not-a-uuid")).toBeUndefined();
		expect(str_u8("1234")).toBeUndefined();
	});
	it("u8_str truncates long buffers (first 16 bytes used)", () => {
		const long = new Uint8Array(20);
		for (let i = 0; i < 20; i++) long[i] = i;
		const s = u8_str(long);
		assert(s !== undefined);

		const u = str_u8(s);
		assert(u !== undefined);
		expect(Array.from(u)).toEqual(Array.from(long.slice(0, 16)));
	});
});
