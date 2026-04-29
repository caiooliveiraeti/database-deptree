# Build stage
FROM golang:1.22.4-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /deptree ./cmd/deptree

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /deptree /usr/local/bin/deptree

ENTRYPOINT ["deptree"]
CMD ["--help"]
