import { resolve } from "node:path";
import { defineConfig } from "vite";

export default defineConfig({
  root: resolve(__dirname, "src/preview"),
  base: "./",
  publicDir: false,
  build: {
    outDir: resolve(__dirname, "dist/__chatcms"),
    emptyOutDir: true,
    rollupOptions: {
      input: {
        select: resolve(__dirname, "src/preview/select.ts"),
        pdf: resolve(__dirname, "src/preview/pdf/index.html"),
      },
      output: {
        entryFileNames: (chunk) => (chunk.name === "select" ? "select.js" : "assets/[name]-[hash].js"),
        chunkFileNames: "assets/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash][extname]",
      },
    },
  },
});
