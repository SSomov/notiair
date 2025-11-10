import adapter from '@sveltejs/adapter-auto';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
		alias: {
			'@components': 'src/lib/components',
			'@stores': 'src/lib/stores',
			'@types': 'src/lib/types',
			'@api': 'src/lib/api'
		}
	}
};

export default config;

