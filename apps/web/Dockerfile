FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /notopia
RUN npm install -g pnpm@10.28.2
COPY . .
RUN pnpm --filter @notopia-uit/web... install --frozen-lockfile && \
  pnpm --filter @notopia-uit/web exec next build

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
