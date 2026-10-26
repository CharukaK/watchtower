FROM node:24-alpine AS assets
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY styles ./styles
RUN npm run css

FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=assets /src/web/static/app.css ./web/static/app.css
RUN go tool templ generate && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /watchtower ./cmd/watchtower

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && adduser -D -H -u 10001 watchtower
COPY --from=build /watchtower /usr/local/bin/watchtower
USER watchtower
EXPOSE 8080
ENTRYPOINT ["watchtower"]
