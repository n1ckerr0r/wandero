import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
      "/health": "http://localhost:8080",
    },
  },
  preview: {
    port: 4173,
  },
  build: {
    sourcemap: false,
    // MapLibre is lazy-loaded as a separate map chunk; its renderer is intentionally larger.
    chunkSizeWarningLimit: 1200,
  },
});
