/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				background: '#f8fafc',
				surface: '#ffffff',
				surfaceMuted: '#f1f5f9',
				border: '#e2e8f0',
				text: '#0f172a',
				muted: '#64748b',
				accent: {
					DEFAULT: '#2563eb',
					foreground: '#ffffff'
				},
				positive: '#0ea5e9',
				warning: '#f59e0b',
				negative: '#ef4444'
			},
			boxShadow: {
				glass: '0 24px 48px -28px rgba(37, 99, 235, 0.25)'
			}
		}
	},
	plugins: []
};
