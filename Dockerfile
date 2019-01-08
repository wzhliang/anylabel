FROM alpine:latest

ADD any-label /any-label
ENTRYPOINT ["./any-label"]
