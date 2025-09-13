# usage: (TODO copy to read me)
# docker build \
#   --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#   -t ss_wallet .

# ---- Builder stage ----
FROM golang:1.23 AS builder

WORKDIR /src

ARG GIT_COMMIT=unknown

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN set -ex && \
    mkdir -p /out && \
    for dir in ./bin/*; do \
        if [ -f "$dir/main.go" ]; then \
            app=$(basename "$dir"); \
            echo "Building $app with commit ${GIT_COMMIT}"; \
            CGO_ENABLED=0 go build \
                -ldflags="-X 'wallet/lib/utils/logger.GitCommit=${GIT_COMMIT}'" \
                -o /out/$app "$dir"; \
        fi \
    done && \
    cp $(go env GOPATH)/bin/goose /out/

# ---- Final stage ----
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /out/ ./bin/
COPY migrations ./migrations

ENV PATH="/app/bin:${PATH}"

