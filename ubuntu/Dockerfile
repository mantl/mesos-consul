FROM ubuntu:15.04

MAINTAINER Chris Aubuchon <Chris.Aubuchon@gmail.com>

RUN apt-get update -y \
	&& apt-get install -y golang git mercurial \
	&& export GOPATH=/go \
	&& go get github.com/CiscoCloud/mesos-consul \
	&& cd /go/src/github.com/CiscoCloud/mesos-consul \
	&& go build -o /bin/mesos-consul \
	&& cd / \
	&& rm -rf /go \
	&& apt-get purge -y golang git mercurial \
	&& apt-get purge -y man perl-modules vim-common vim-tiny \
		libpython3.4-stdlib:amd64 python3.4-minimal xkb-data \
		libx11-data eject python3 locales golang-go \
	&& apt-get clean autoclean \
	&& apt-get autoremove -y \
	&& rm -rf /var/lib/{apt,dpkg,cache,log}/

ENTRYPOINT [ "/bin/mesos/consul" ]
