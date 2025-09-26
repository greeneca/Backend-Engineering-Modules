# Build stage
FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./
COPY configuration ./configuration
COPY server ./server
COPY statistics ./statistics
COPY wiki ./wiki

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -o wiki_updates

# Runtime stage
FROM gcr.io/distroless/base

COPY --from=builder /app/wiki_updates /wiki_updates
COPY wiki_updates.conf.json /wiki_updates.conf.json

ENTRYPOINT ["/wiki_updates"]
