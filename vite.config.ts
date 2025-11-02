import path, { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import dts from "vite-plugin-dts";
import { defineConfig } from "vitest/config";

const __dirname = dirname(fileURLToPath(import.meta.url));

// https://vite.dev/config/
export default defineConfig({
	build: {
		minify: false,
		sourcemap: true,
		lib: {
			entry: {
				index: resolve(__dirname, "src/index.ts"),
			},
			formats: ["es"],
			fileName: (format, entryName) => `${entryName}.${format}.js`,
		},
		rollupOptions: {
			external: [
				"@bufbuild/protobuf",
				"@connectrpc/connect",
				"dexie",
			],
		},
	},
	plugins: [
		dts({
			tsconfigPath: "./tsconfig.json",
			exclude: [
				"examples",
				"src/**/*.test.ts",
			],
		}),
	],
	test: {},
	resolve: {
		alias: {
		"@/gen": path.resolve(__dirname, "./examples/gen"),
		"@protobuf-orm/runtime": path.resolve(__dirname, "./src/index.ts"),
		}
	}
});
