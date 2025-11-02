/**
 * Convert a 16-byte Uint8Array into a UUID string (8-4-4-4-12).
 *
 * Notes:
 * - If `v` is longer than 16 bytes, only the first 16 bytes are used.
 * - If `v` is not provided or has length other than 16, the function returns `undefined`.
 * - The returned string uses lowercase hex (e.g. "550e8400-e29b-41d4-a716-446655440000").
 *
 * Example:
 *   u8_str(new Uint8Array([0x55,0x0e,0x84,0x00,...])) -> "550e8400-..."
 */
export function u8_str(v?: Uint8Array): string | undefined {
	v = v?.slice(0, 16)
	if (v?.length !== 16) return undefined

	// build 32-char hex string
	let hex = ''
	for (let i = 0; i < 16; i++) {
		const b = v[i]
		const h = b.toString(16)
		hex += h.length === 1 ? `0${h}` : h
	}

	// UUID format: 8-4-4-4-12
	return (
		hex.slice(0, 8) +
		'-' +
		hex.slice(8, 12) +
		'-' +
		hex.slice(12, 16) +
		'-' +
		hex.slice(16, 20) +
		'-' +
		hex.slice(20)
	)
}

/**
 * Parse a UUID string into a 16-byte Uint8Array.
 *
 * Behavior and accepted input:
 * - Any non-hex characters are stripped before parsing, so the function accepts
 *   standard forms like `550e8400-e29b-41d4-a716-446655440000`, `{...}`, or
 *   `urn:uuid:...`.
 * - After stripping, exactly 32 hex characters must remain; otherwise `undefined`
 *   is returned.
 * - On any parse error the function returns `undefined`.
 *
 * Example:
 *   str_u8("550e8400-e29b-41d4-a716-446655440000") -> Uint8Array(16)
 */
export function str_u8(v?: string): Uint8Array | undefined {
	if (!v) return undefined

	// Remove all non-hex characters (dashes, braces, etc.)
	const s = v.replace(/[^0-9a-fA-F]/g, '').slice(0, 32)
	if (s.length !== 32) return undefined

	const out = new Uint8Array(16)
	for (let i = 0; i < 32; i += 2) {
		const byteHex = s.slice(i, i + 2)
		const parsed = parseInt(byteHex, 16)
		if (Number.isNaN(parsed)) return undefined
		out[i >> 1] = parsed
	}
	return out
}
