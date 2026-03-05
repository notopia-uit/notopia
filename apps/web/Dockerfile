FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /notopia
RUN npm install -g pnpm@10.28.2
COPY . .
RUN pnpm --filter @notopia-uit/web... install --frozen-lockfile && \
  pnpm exec nx build web

FROM node:25.6.1-alpine3.23 AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=builder /notopia/apps/web/public ./public
COPY --from=builder /notopia/apps/web/.next/standalone ./
COPY --from=builder /notopia/apps/web/.next/static ./.next/static
COPY --from=builder /notopia/apps/web/.next/cache ./.next/cache
EXPOSE 3000
CMD ["node", "server.js"]
