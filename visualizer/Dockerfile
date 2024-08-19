FROM node:latest AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci
COPY . ./
RUN echo "" > ./src/routes/dev-interceptor.ts
RUN rm ./src/hooks.server.ts
RUN npm run build

FROM node:lts-alpine3.20

RUN apk update && apk upgrade && apk add dumb-init
USER 1000

# Copy node_modules from the build stage
COPY --chown=1000:1000 --from=build /app/node_modules ./node_modules
COPY --chown=1000:1000 --from=build /app/build ./build
COPY --chown=1000:1000 --from=build /app/package.json ./
COPY --chown=1000:1000 --from=build /app/package-lock.json ./

EXPOSE 8080

ENV HOST=0.0.0.0 PORT=8080 NODE_ENV=production

CMD ["dumb-init","node","build/index.js"]
