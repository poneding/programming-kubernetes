FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY examples/admission-controller/main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o controller .

FROM alpine:latest
COPY --from=builder /app/controller /controller

EXPOSE 443
ENTRYPOINT ["/controller"]
