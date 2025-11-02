function decay(v?: Record<string, unknown>): Record<string, unknown> {
	return v ?? {};
}

export function rm(obj: Record<string, unknown>, key: string) {
	delete obj[key];
}

export function set<T extends {}, V>(
	obj: T | undefined,
	key: keyof T,
	value: V,
) {
	decay(obj)[key as string] = value;
}
