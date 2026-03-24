FROM golang:1.22-bookworm AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /prompt-evolver .
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates git && rm -rf /var/lib/apt/lists/*
COPY --from=builder /prompt-evolver /usr/local/bin/prompt-evolver
ENTRYPOINT ["prompt-evolver"]
