FROM alpine:latest
ARG TARGETARCH

WORKDIR /run

COPY ./bin/$TARGETARCH/controller ./controller

ENTRYPOINT ["./controller"]
