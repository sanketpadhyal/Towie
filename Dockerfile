# Stage 1 — Build
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache dependencies before copying source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X github.com/sanketpadhyal/towie/internal/buildinfo.Version=${VERSION} \
      -X github.com/sanketpadhyal/towie/internal/buildinfo.Commit=${COMMIT} \
      -X github.com/sanketpadhyal/towie/internal/buildinfo.Date=${DATE}" \
    -o /bin/towie \
    ./cmd/towie

# Stage 2 — Runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/towie /towie

EXPOSE 8080

ENTRYPOINT ["/towie", "start"]
