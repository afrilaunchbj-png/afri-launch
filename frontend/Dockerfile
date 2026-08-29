# syntax=docker/dockerfile:1

# ---------- Build SPA ----------
FROM node:22-alpine AS build
WORKDIR /app
RUN npm install -g pnpm@11.14.0
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --config.dangerouslyAllowAllBuilds=true
COPY . .

ARG VITE_API_URL=""
ARG VITE_NEON_AUTH_URL=""
ENV VITE_API_URL=$VITE_API_URL VITE_NEON_AUTH_URL=$VITE_NEON_AUTH_URL

RUN pnpm build

# ---------- Serve (nginx + SPA fallback) ----------
FROM nginx:alpine
ENV PORT=8080
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
