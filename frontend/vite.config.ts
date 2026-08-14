import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "https://42plusplus-team.accelolabs.com",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://42plusplus-team.accelolabs.com",
        ws: true,
      },
    },
  },
});
