# Web frontend builder
FROM node:26-alpine AS web-builder
WORKDIR /app/web

RUN npm install -g corepack@latest && \
    corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY web ./
RUN pnpm run build

# Main go builder
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
ARG TARGETPLATFORM

RUN apk add --no-cache git build-base tzdata

WORKDIR /app
# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
COPY --from=web-builder /app/web/build.go ./web

# CGO/build environments
ENV CGO_ENABLED=0

RUN go build -a -v \
    -trimpath \
    --installsuffix cgo \
    -ldflags="-s -w -extldflags '-static'" \
    -o app ./

## deploy
FROM alpine
WORKDIR /

RUN apk --no-cache add ca-certificates tzdata

USER nonroot

COPY --from=builder /app/configs /configs
COPY --from=builder /app/app /app



ENTRYPOINT ["/app"]