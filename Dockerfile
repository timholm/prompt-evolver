FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /prompt-evolver .

FROM alpine:3.20

RUN apk add --no-cache sqlite-libs git ca-certificates

COPY --from=builder /prompt-evolver /usr/local/bin/prompt-evolver

ENTRYPOINT ["prompt-evolver"]
