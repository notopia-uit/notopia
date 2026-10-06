FROM node:25.6.1-alpine3.23 AS builder
ENV NODE_ENV=production
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
WORKDIR /app
RUN npm install -g pnpm@10.33.2
COPY apps/ai/dist/ .
RUN pnpm install --frozen-lockfile --prod
EXPOSE 8085
CMD ["node", "main.cjs"]
