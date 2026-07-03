# ============================================================
# Dockerfile — HRMS Frontend (SvelteKit)
# Multi-stage: development (HMR) + production (Node server)
# ============================================================

# ---- Stage 1: Base ----
FROM node:22-alpine AS base
WORKDIR /app

# ---- Stage 2: Development (HMR with polling) ----
FROM base AS dev

# Copy package files
COPY frontend/package*.json ./
COPY frontend/.npmrc ./

RUN npm install

# Copy all source
COPY frontend/ .

EXPOSE 5173

ENV HOST=0.0.0.0
ENV PORT=5173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# ---- Stage 3: Build for production ----
FROM base AS builder

COPY frontend/package*.json ./
COPY frontend/.npmrc ./
RUN npm install

COPY frontend/ .

RUN npm run build

# ---- Stage 4: Production (static files via Nginx) ----
FROM nginx:alpine AS prod

COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx config for SPA fallback with Brotli + gzip compression
RUN echo 'server { \
    listen 80; \
    root /usr/share/nginx/html; \
    index index.html; \
\n    # Gzip compression\n    gzip on;\n    gzip_comp_level 5;\n    gzip_min_length 256;\n    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript image/svg+xml;\n    gzip_vary on;\n    gzip_proxied any;\n    gzip_disable "msie6";\n\n    # Cache static assets (immutable chunks)\n    location /_app/immutable/ {\n        expires 1y;\n        add_header Cache-Control "public, immutable, max-age=31536000";\n        add_header X-Content-Type-Options nosniff;\n        try_files $uri $uri/ /index.html;\n    }\n\n    # Cache other static assets\n    location /_app/ {\n        expires 30d;\n        add_header Cache-Control "public, max-age=2592000";\n        try_files $uri $uri/ /index.html;\n    }\n\n    # SPA fallback — semua route ke index.html\n    location / {\n        expires -1;\n        add_header Cache-Control "no-cache, must-revalidate";\n        try_files $uri $uri/ /index.html;\n    }\n\n    # API proxy ke backend\n    location /api/ {\n        proxy_pass http://api:8080;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n}' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
