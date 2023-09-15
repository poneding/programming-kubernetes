#!/bin/bash

openapi-gen -i ./... -o ./openapi_generated.go

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/amd64/controller main.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/arm64/controller main.go

docker buildx use pkbuilder || docker buildx create --use --name pkbuilder
docker buildx build --push --platform linux/amd64,linux/arm64 -t poneding/echo-hello-sidecar-admission-controller:latest .

