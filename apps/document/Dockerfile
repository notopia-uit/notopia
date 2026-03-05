FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
WORKDIR /notopia
RUN npm install -g pnpm@10.28.2
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile --filter document && \
  pnpm build --filter document && \
  pnpm deploy --filter document --prod /app

FROM node:25.6.1-alpine3.23 AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/ ./
EXPOSE 8082
CMD ["node", "main.js"]
