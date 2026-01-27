import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import basicSsl from "@vitejs/plugin-basic-ssl";
import dotenv from "dotenv";
import path from "path";
import fs from "fs";

const environmentFiles = ["../.env.development", "../.env.production"];

for (const fileName in environmentFiles) {
	const filePath = path.resolve(__dirname, environmentFiles[fileName]);
	if (fs.existsSync(filePath)) {
		dotenv.config({
			path: [filePath, "../.env"],
		});
		break;
	}
}

let targetHostAddress = process.env.HTTP_ADDRESS || "localhost:8080";
if (targetHostAddress[0] === ":") {
	targetHostAddress = "localhost" + targetHostAddress;
}

let targetProtocol = "http://";
targetHostAddress = "localhost:80";

if (process.env.USE_SSL == "TRUE") {
	const httpsPort = "443";
	targetProtocol = "https://";

	const currentTarget = targetHostAddress.split(":");
	if (currentTarget.length === 1) {
		targetHostAddress += httpsPort;
	} else {
		targetHostAddress = targetHostAddress.replace(currentTarget[1], httpsPort);
	}
}

console.log(`Target Backend: ${targetProtocol}${targetHostAddress}`);

export default defineConfig({
	plugins: [react(), tailwindcss(), basicSsl()],
	server: {
		host: targetHostAddress,
		https: true,
		open: true,
		proxy: {
			"/api": {
				target: `${targetProtocol}${targetHostAddress}`,
				changeOrigin: true,
				secure: false,

				configure: (proxy, _) => {
					proxy.on("proxyReq", (proxyReq, req, _) => {
						if (req.url != "/api/status") {
							console.log("----------------------------");
							console.log("🔹 Request Origin:", req.headers.origin);
							console.log("🔹 Request Method:", req.method);
							console.log("🔹 Request URL:", req.url);
							console.log("----------------------------");
							console.log("");
						}

						if (req.headers.accept === "text/event-stream") {
							proxyReq.setHeader("Connection", "keep-alive");
							proxyReq.setHeader("Cache-Control", "no-cache");
							proxyReq.setHeader("X-Accel-Buffering", "no");
						}
					});
				},
			},
		},
	},
	css: {
		postcss: "./postcss.config.js",
	},
	build: {
		outDir: "build",
		rollupOptions: {
			output: {
				manualChunks(id) {
					if (id.includes('node_modules')) {
						return "node"
					}
					if (id.includes('src/components')) {
						return "components"
					}
				}
			}
		}
	},
	envDir: "../",
	// For backwards compatibility
	envPrefix: ["REACT_", "VITE_"],
});
