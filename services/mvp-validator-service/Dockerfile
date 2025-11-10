FROM node:20-alpine
WORKDIR /app
COPY package.json pnpm-lock.yaml tsconfig.json ./
RUN corepack enable && pnpm i --frozen-lockfile
COPY src ./src
EXPOSE 7402
CMD ["pnpm","start"]
