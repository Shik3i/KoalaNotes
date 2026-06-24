import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import adapter from '@sveltejs/adapter-static';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({
				// SPA mode for offline-first: fallback to index.html for all routes
				fallback: 'index.html',
				pages: 'build',
				assets: 'build',
				precompress: false,
				strict: true
			})
		})
	]
});
