FROM alpine:3.5

MAINTAINER Chris Aubuchon <Chris.Aubuchon@gmail.com>

COPY . /mesos-consul-source
RUN apk add --update gcc g++ go git mercurial \
	&& mkdir -p /go/src/github.com/CiscoCloud \
	&& cp -a /mesos-consul-source /go/src/github.com/CiscoCloud/mesos-consul \
	&& cd /go/src/github.com/CiscoCloud/mesos-consul \
	&& export GOPATH=/go \
	&& go get \
	&& go build -o /bin/mesos-consul \
	&& apk del --purge gcc g++ go git mercurial \
	&& rm -rf /go

ENTRYPOINT [ "/bin/mesos-consul" ]
