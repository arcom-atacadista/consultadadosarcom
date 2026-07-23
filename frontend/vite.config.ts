import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";

// No Docker: API_TARGET=http://backend:3000 (nome do serviço no compose).
// Local (pnpm dev, fora do Docker): cai no localhost:3000.
const apiTarget = process.env.API_TARGET || "http://localhost:3000";
const proxy = { "/api": { target: apiTarget, changeOrigin: true } };

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      manifest: {
        name: "CDA — Consulta Dados Arcom",
        short_name: "CDA",
        theme_color: "#007840",
      },
    }),
  ],
  resolve: { alias: { "@": "/src" } },
  server: { host: true, port: 5173, proxy },   // pnpm dev
  preview: { host: true, port: 4173, proxy },  // pnpm preview (usado no Docker)
});
