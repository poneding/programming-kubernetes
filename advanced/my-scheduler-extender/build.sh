GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/linux/amd64/my-scheduler-extender
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/linux/arm64/my-scheduler-extender

docker buildx build --push --platform linux/amd64,linux/arm64 -t poneding/my-kube-scheduler-extender:v1.0 .
