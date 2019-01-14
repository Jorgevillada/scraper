FROM golang:1.10.3-alpine3.8 as builder
LABEL maintainer="Jorge Villada"

ENV PORT=3737
# Commented-out because these are defaults anyway
# ENV GOPATH=/go
# ENV PATH=${GOPATH}/bin:${PATH}
ENV APP_USER=appuser
ENV SRC_PATH=/go/src/craw4-scraper
ENV APP_ENV=production

COPY . ${SRC_PATH}
WORKDIR ${SRC_PATH}

USER root

RUN adduser -s /bin/false -D ${APP_USER} \
 && echo "Installing git and bash support" \
 && apk update && apk upgrade \
 && apk add --no-cache bash git \
 && echo "Installing infrastructural go packages…" \
 && go get -u github.com/pilu/fresh \
 && go get -u github.com/golang/dep/cmd/dep \
 && echo "Installing Dependencies…" \
 && goWrapProvision="$(go-wrapper fake 2>/dev/null || true)" \
 && cd ${SRC_PATH} && dep ensure && dep ensure -update \
 && echo "building project..." \
 && go build \
 && echo "Fixing permissions..." \
 && chown -R ${APP_USER}:${APP_USER} ${GOPATH} \
 && chown -R ${APP_USER}:${APP_USER} ${SRC_PATH} \
 && chmod u+x ${SRC_PATH}/scripts/*.sh \
 && echo "Cleaning up installation caches to reduce image size" \
 && rm -rf /root/src /tmp/* /usr/share/man /var/cache/apk/*
USER ${APP_USER}

EXPOSE ${PORT}




FROM alpine as release
ENV APP_ENV=production
ENV PORT=3737

WORKDIR /
 RUN  echo "Installing git and bash support" \
 && apk update && apk upgrade \
 && apk add ca-certificates \
 && echo "Cleaning up installation caches to reduce image size" \
 && rm -rf /root/src /tmp/* /usr/share/man /var/cache/apk/*
COPY --from=builder /go/src/craw4-scraper/craw4-scraper .
CMD ["/craw4-scraper"]
