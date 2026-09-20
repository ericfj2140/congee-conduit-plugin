import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const rootDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	plugins: [tailwindcss(), svelte()],
	base: './',
	server: {
		fs: { allow: [path.resolve(rootDir, '..')] }
	},
	build: {
		outDir: '../ui',
		emptyOutDir: true
	}
});
