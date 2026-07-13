# ========== 阶段1：打包前端 ==========
FROM node:22-alpine AS builder

WORKDIR /app
COPY amap-uni/package.json amap-uni/package-lock.json* ./
RUN npm ci

COPY amap-uni/ ./
RUN npm run build:server

# ========== 阶段2：Nginx 托管 ==========
FROM nginx:1.27-alpine

COPY amap-uni/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80