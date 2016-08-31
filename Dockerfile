FROM alpine:3.3

RUN apk update \
    && apk add --no-cache openssh ca-certificates \
    && rm -rf /var/cache/apk/*

ADD pkg/linux_amd64/slackstack /usr/bin/slackstack

ENTRYPOINT ["slackstack"]