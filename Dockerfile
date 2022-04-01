FROM golang:1.17-bullseye as build
ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -a -o /rcdp

FROM chromedp/headless-shell:latest as chrome-headless

WORKDIR /

COPY docker/go-debug/root/headless.sh /
COPY docker/go-debug/root/runner.sh /
COPY --from=build /rcdp /
RUN mkdir docs
COPY ./docs/swagger.* ./docs/

USER root

RUN apt-get update -y && \
    apt-get install -y --no-install-recommends \
      apt-transport-https \
      ca-certificates \
      dumb-init \
      apt-utils && \
    apt-get autoremove -y --purge && \
    apt-get autoclean -y && \
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    chmod -R 777 /tmp && \
    chmod -R  777 /var/tmp && \
    chmod a+x /headless.sh && \
    chmod a+x /runner.sh && \
    chmod a+x /rcdp

ENTRYPOINT ["dumb-init", "--"]

FROM chrome-headless as rcdp
LABEL name="remote-chrome-devtools-protocol" \
	  maintainer="Elton Schivei Costa <costa@elton.schivei.nom.br>" \
	  version="1.0" \
	  description="Remote Google Chrome DevTools Protocol"

VOLUME /tmp
VOLUME /var/tmp
EXPOSE 12345
EXPOSE 9222

CMD /runner.sh
