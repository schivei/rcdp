FROM chromedp/headless-shell:latest as chrome-headless
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
        apt-transport-https \
        ca-certificates \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
        wget \
        libnspr4 \
        libnss3 \
        libexpat1 \
        libfontconfig1 \
        libuuid1 \
        dpkg \
        git \
        musl-dev \
        dumb-init \
        curl \
        parallel \
	; \
	rm -rf /var/lib/apt/lists/*
ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION 1.18

WORKDIR /tmp

RUN wget https://golang.org/dl/go1.18.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH" \
    && rm -rf /tmp/* \
    && chmod -R 777 /tmp

RUN go version

WORKDIR $GOPATH

FROM chrome-headless

COPY root /

RUN chmod a+x /entrypoint.sh && \
    chmod a+x /headless.sh && \
    chmod a+x /debugger.sh

ENV GO111MODULE auto
ENV CGO_ENABLED 1
ENV TEMP /tmp

RUN go get -u github.com/go-delve/delve/cmd/dlv

USER root

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 2345
EXPOSE 9222
EXPOSE 12345

CMD ["parallel", "--ungroup", ":::", "/headless.sh", "/debugger.sh", ":::"]
