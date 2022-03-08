FROM golang:1.17-alpine AS backend-builder

WORKDIR /app

COPY backend/go.mod .
COPY backend/go.sum .
RUN go mod download

ADD backend/ .

RUN go build -o backend .

FROM node:16 AS frontend-builder

ENV NODE_ENV=production

WORKDIR /app

COPY frontend/package.json .
COPY frontend/package-lock.json* .

RUN npm install --production

ADD frontend/ .

RUN npm run build

FROM alpine

COPY --from=frontend-builder /app/dist/ /dist/
COPY --from=backend-builder /app/backend /
COPY --from=backend-builder /app/config.yaml /

CMD ["/backend", "-config", "/config.yaml"]
