# syntax=docker/dockerfile:1

FROM node:20-alpine
WORKDIR /app

COPY assignment_vue/frontend-vue/package*.json ./
RUN npm ci

COPY assignment_vue/frontend-vue/ ./

EXPOSE 5173
CMD ["sh", "-c", "npm run dev -- --host ${FRONTEND_HOST:-0.0.0.0} --port ${FRONTEND_PORT:-5173}"]
