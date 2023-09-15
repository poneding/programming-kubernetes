#!/bin/bash

openapi-gen \
  -i pk/examples/my-apiserver \
  -p pk/examples/my-apiserver/generated \
  -o . \
  --output-file-base zz_generated.openapi \
  --trim-path-prefix pk/examples/my-apiserver \
  --go-header-file ./boilerplate.go.txt

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/amd64/my-apiserver main.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/arm64/my-apiserver main.go

docker buildx use pkbuilder || docker buildx create --use --name pkbuilder
docker buildx build --push --platform linux/amd64,linux/arm64 -t registry.cn-hangzhou.aliyuncs.com/pding/my-apiserver:latest .

rm -rf bin/