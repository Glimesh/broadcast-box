import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
	plugins: [react(), tailwindcss()],
	css: {
		postcss: './postcss.config.js',
	},
	build: {
		outDir: 'build',
	},
	server: {
		open: true, // Opens browser on dev server start
	},
	envDir: '../',
});
