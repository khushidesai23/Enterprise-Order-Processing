import { defineConfig, loadEnv, type ProxyOptions } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "node:path";

// The browser talks to /api on its own origin; Vite forwards it to the Go
// API, so no CORS setup is needed in development or `vite preview`.
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiTarget = env.API_PROXY_TARGET || "http://localhost:8080";
  const proxy: Record<string, ProxyOptions> = {
    "/api": {
      target: apiTarget,
      changeOrigin: true,
      // The browser only talks to this server, so the API needs no CORS
      // check; dropping Origin keeps any dev/preview port working.
      configure: (server) => server.on("proxyReq", (request) => request.removeHeader("origin")),
    },
  };

  return {
    plugins: [react(), tailwindcss()],
    resolve: { alias: { "@": path.resolve(__dirname, "src") } },
    server: { port: 5173, proxy },
    preview: { port: 4173, proxy },
    test: {
      environment: "jsdom",
      include: ["src/**/*.test.{ts,tsx}"],
    },
  };
});
