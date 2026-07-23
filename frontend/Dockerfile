# Faz o build e serve o resultado com o "vite preview".
# Sem nginx, sem Traefik: o proxy de /api pro backend está no vite.config.ts.
FROM node:20-alpine
WORKDIR /app
RUN corepack enable
COPY .npmrc package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build
EXPOSE 4173
CMD ["pnpm", "run", "preview", "--host", "0.0.0.0", "--port", "4173"]
