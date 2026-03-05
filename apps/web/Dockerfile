FROM node:25.6.1-slim AS builder
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN corepack enable pnpm
WORKDIR /notopia
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile --filter web && \
  pnpm build --filter web && \
  pnpm deploy --filter web --prod /app

FROM node:25.6.1-alpine3.23 AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=builder /app/public ./public
RUN mkdir .next
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/.next/cache ./.next/cache
EXPOSE 3000
CMD ["node", "server.js"]
