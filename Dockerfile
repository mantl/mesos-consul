FROM gliderlabs/alpine:3.1

MAINTAINER Chris Aubuchon <Chris.Aubuchon@gmail.com>
ARG http_proxy=http://proxy.expert.de:8080
ARG https_proxy=http://proxy.expert.de:8080


COPY . /go/src/github.com/CiscoCloud/mesos-consul
RUN apk add --update go git mercurial \
	&& cd /go/src/github.com/CiscoCloud/mesos-consul \
	&& export GOPATH=/go \
	&& go get \
	&& go build -o /bin/mesos-consul \
	&& rm -rf /go \
	&& apk del --purge go git mercurial

ENTRYPOINT [ "/bin/mesos-consul" ]
