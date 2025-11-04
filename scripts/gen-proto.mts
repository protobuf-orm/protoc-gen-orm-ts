#!/usr/bin/env -S npx tsx
import "zx/globals";

export const rootDir = path.join(path.dirname(import.meta.dirname), 'examples')

const service_files = await glob([
	path.join(rootDir, "proto.svc/**/*.g.proto"),
]);
for (const f of service_files) {
	const d = path.dirname(f);
	const n = path.basename(f, ".g.proto");
	const p = path.join(d, `${n}.ext.proto`);

	const r = path.relative(path.join(rootDir, "proto.svc"), d);
	const v = path.join(path.join(rootDir, "proto", r), `${n}.g.proto`);
	if (fs.existsSync(p)) {
		const o = fs.createWriteStream(v);
		await $`go tool github.com/lesomnus/proto-merge ${f} ${p}`.pipe(o);
	} else {
		fs.copyFileSync(f, v);
	}
}
