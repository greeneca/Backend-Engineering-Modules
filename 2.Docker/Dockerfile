FROM golang:1.25.1

WORKDIR /app

RUN mkdir wiki statstics server

COPY go.mod go.sum main.go ./
COPY server/ server/
COPY wiki/ wiki/
COPY statstics/ statstics/

RUN CGO_ENABLED=0 GOOS=linux go build -o /wiki_stats

ENTRYPOINT ["/wiki_stats"]
