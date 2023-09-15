FROM alpine
ARG TARGETARCH
WORKDIR /run
COPY ./bin/$TARGETARCH/my-apiserver ./my-apiserver

ENTRYPOINT ["./my-apiserver"]