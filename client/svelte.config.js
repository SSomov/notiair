import adapter from "@sveltejs/adapter-node";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

const config = {
	preprocess: vitePreprocess(),
	kit: {
		paths: {
			base: process.env.NOTIAIR_BASE_PATH ?? "",
		},
		adapter: adapter(),
		alias: {
			"@components": "src/lib/components",
			"@stores": "src/lib/stores",
			"@types": "src/lib/types",
			"@api": "src/lib/api",
		},
	},
};

export default config;
