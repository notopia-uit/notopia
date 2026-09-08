FROM node:25.9.0-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
WORKDIR /app
RUN npm install -g pnpm@10.33.2
COPY apps/search-worker/dist/ .
RUN pnpm install --frozen-lockfile --prod
EXPOSE 8083
CMD ["node", "main.js"]
