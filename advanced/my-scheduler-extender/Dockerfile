FROM alpine
ARG TARGETOS TARGETARCH
ADD ./bin/$TARGETOS/$TARGETARCH/my-scheduler-extender /my-scheduler-extender
ENTRYPOINT ["/my-scheduler-extender"]