# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o gose ./cmd

FROM node:20 AS frontend-builder

WORKDIR /app

COPY frontend/package.json .
COPY frontend/package-lock.json* .

RUN npm install

COPY frontend/ .

ENV NODE_ENV=production

RUN npm run build

FROM alpine:3.18

RUN apk update && apk add ca-certificates curl && rm -rf /var/cache/apk/*

COPY --from=frontend-builder /app/dist/ /dist/
COPY --from=backend-builder /app/gose /
COPY --from=backend-builder /app/config.yaml /

ENV GIN_MODE=release
ENV GOSE_SERVER_STATIC=/dist

EXPOSE 8080/tcp

HEALTHCHECK --interval=30s --timeout=30s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/healthz

ENTRYPOINT [ "/gose" ]
