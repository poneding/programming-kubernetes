FROM alpine
ARG TARGETOS TARGETARCH
ADD ./bin/$TARGETOS/$TARGETARCH/my-scheduler-controller /my-scheduler-controller
ENTRYPOINT ["/my-scheduler-controller"]