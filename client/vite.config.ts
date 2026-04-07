import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: process.env.PORT ? Number(process.env.PORT) : 5173,
		// Проксирование в dev режиме (для SSR=false)
		// В production hooks.server.ts обрабатывает проксирование
		proxy:
			process.env.NODE_ENV === "production"
				? {}
				: {
						"/api": {
							target: process.env.PUBLIC_API_URL || "http://127.0.0.1:8080",
							changeOrigin: true,
						},
						"/sub": {
							target: process.env.PUBLIC_API_URL || "http://127.0.0.1:8080",
							changeOrigin: true,
						},
					},
	},
});
