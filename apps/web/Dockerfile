FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /app
COPY apps/web/public ./public
COPY apps/web/.next/standalone ./
COPY apps/web/.next/static ./.next/static
COPY apps/web/.next/cache ./.next/cache
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
EXPOSE 3000
ENTRYPOINT ["node", "apps/web/server.js"]
