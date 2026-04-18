# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM node:22-alpine AS builder

WORKDIR /app

# Install dependencies first (cached layer)
COPY package*.json ./
RUN npm ci

# Copy source
COPY . .

# VITE_API_BASE_URL is baked into the JS bundle at build time.
# docker-compose passes http://localhost:8080 so the browser can reach
# the backend container which is exposed on the host at port 8080.
ARG VITE_API_BASE_URL=http://localhost:8080
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM nginx:stable-alpine

# Copy the compiled SPA from the build stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Replace the default nginx site config with ours (handles SPA routing)
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
