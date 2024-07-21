FROM node:latest AS build

WORKDIR /app

COPY package.json ./
COPY package-lock.json ./
RUN npm ci
COPY . ./
RUN npm run build

FROM node:lts-alpine3.20

RUN apk update && apk upgrade && apk add dumb-init
USER 1000

WORKDIR /app

COPY --chown=1000:1000 --from=build /app/build /app/package.json ./

EXPOSE 8080

ENV HOST=0.0.0.0 PORT=8080 NODE_ENV=production

CMD ["dumb-init","node","index.js"]