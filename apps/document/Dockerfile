FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
WORKDIR /notopia
RUN npm install -g pnpm@10.28.2
COPY . .
RUN pnpm --filter @notopia-uit/document... install --frozen-lockfile && \
  pnpm exec nx build document && \
  pnpm --filter @notopia-uit/document deploy --prod /app --legacy

FROM node:25.6.1-alpine3.23 AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NODE_OPTIONS="--require @opentelemetry/auto-instrumentations-node/register"
COPY --from=builder /app/ ./
EXPOSE 8082
CMD ["node", "dist/main.js"]
