FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
WORKDIR /notopia
RUN npm install -g pnpm@10.28.2
COPY . .
RUN pnpm install --filter @notopia-uit/document --frozen-lockfile && \
  pnpm exec --filter @notopia-uit/document webpack-cli build && \
  pnpm deploy --filter @notopia-uit/document --prod /app

FROM node:25.6.1-alpine3.23 AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/ ./
EXPOSE 8082
CMD ["node", "main.js"]
