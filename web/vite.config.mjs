import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import basicSsl from '@vitejs/plugin-basic-ssl'
import dotenv from 'dotenv'
import path from 'path'

// Use the Go environment file to load variables
dotenv.config(
	{
		// First found .env file is selected
		path: [
			path.resolve(__dirname, '../.env')
			path.resolve(__dirname, '../.env.production')
			path.resolve(__dirname, '../.env.development')
		]
	})

let targetHostAddress = process.env.HTTP_ADDRESS || 'localhost';
let targetHostPort = process.env.HTTP_PORT || '8080';
let targetProtocol = "http://"

if(process.env.USE_SSL == "TRUE"){
	targetHostPort = process.env.HTTPS_PORT || '443';
	targetProtocol = "https://"
}

export default defineConfig({
	plugins: [react(), tailwindcss(), basicSsl()],
	server: {
		host: targetHostAddress,
		https: true,
		open: true,
		proxy: {
			'/api': {
				target: `${targetProtocol}${targetHostAddress}:${targetHostPort}`,
				changeOrigin: true,
				secure: false,

				configure: (proxy, _) => {
					proxy.on('proxyReq', (proxyReq, req, _) => {

						if (req.url != '/api/status') {
							console.log('🔹 Request Origin:', req.headers.origin);
							console.log('🔹 Request Method:', req.method);
							console.log('🔹 Request URL:', req.url);
						}

						if (req.headers.accept === 'text/event-stream') {
							proxyReq.setHeader('Connection', 'keep-alive')
							proxyReq.setHeader('Cache-Control', 'no-cache')
							proxyReq.setHeader('X-Accel-Buffering', 'no')
						}
					})
				}
			}
		}
	},
	css: {
		postcss: './postcss.config.js',
	},
	build: {
		outDir: 'build',
	},
	envDir: '../',
	// For backwards compatibility
	envPrefix: ['REACT_', 'VITE_'],
});
