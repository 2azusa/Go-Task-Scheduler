FROM golang:alpine AS builder

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY
ENV HTTP_PROXY=$HTTP_PROXY
ENV HTTPS_PROXY=$HTTPS_PROXY
ENV NO_PROXY=$NO_PROXY

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o /app/bin/pulsenode ./node/cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/pulsenode .
COPY node/conf ./node/conf

ENTRYPOINT ["/app/pulsenode"]