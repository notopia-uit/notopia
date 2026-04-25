FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /app
RUN npm install -g pnpm@10.33.2
COPY apps/web/public ./public
COPY apps/web/.next/standalone ./
COPY apps/web/.next/static ./.next/static
COPY apps/web/.next/cache ./.next/cache
COPY apps/web/.next/workspace-module/ .
COPY apps/web/.next/package.json/ .
COPY apps/web/.next/pnpm-lock.yaml/ .
RUN pnpm install --frozen-lockfile --prod
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
EXPOSE 3000
CMD ["node", "apps/web/server.js"]
