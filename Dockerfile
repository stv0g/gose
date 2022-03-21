FROM golang:1.17-alpine AS backend-builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o gose ./cmd

FROM node:17 AS frontend-builder

ENV NODE_ENV=production

WORKDIR /app

COPY frontend/package.json .
COPY frontend/package-lock.json* .

RUN npm install --production

COPY frontend/ .

RUN npm run build

FROM alpine:3.15

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=frontend-builder /app/dist/ /dist/
COPY --from=backend-builder /app/gose /
COPY --from=backend-builder /app/config.yaml /

ENV GIN_MODE=release
ENV GOSE_SERVER_STATIC=/dist

ENTRYPOINT [ "/gose" ]
