FROM ubuntu:14.04
MAINTAINER Rafik Salama <rafik@oysterbooks.com>

WORKDIR /opt/go/src/github.com/oysterbooks/halfshell
ENV GOPATH /opt/go

RUN apt-get update && apt-get install -qy \
    build-essential \
    git \
    wget \
    libmagickcore-dev \
    libmagickwand-dev \
    imagemagick \
    golang

ADD . /opt/go/src/github.com/oysterbooks/halfshell
RUN cd /opt/go/src/github.com/oysterbooks/halfshell && make deps && make build

ENTRYPOINT ["/opt/go/src/github.com/oysterbooks/halfshell/bin/halfshell"]

EXPOSE 8080
