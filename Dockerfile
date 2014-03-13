FROM ubuntu
MAINTAINER Rafik Salama <rafik@oysterbooks.com>

WORKDIR /opt/go/src/github.com/oysterbooks/halfshell
ENV GOPATH /opt/go

RUN echo "deb http://archive.ubuntu.com/ubuntu precise main universe" > /etc/apt/sources.list
RUN apt-get update
RUN apt-get upgrade -qy
RUN apt-get install -qy \
    python-software-properties \
    libmagickwand-dev \
    git

RUN add-apt-repository ppa:duh/golang
RUN apt-get update
RUN apt-get install -y golang

ADD . /opt/go/src/github.com/oysterbooks/halfshell
RUN cd /opt/go/src/github.com/oysterbooks/halfshell && make deps && make build

ENTRYPOINT ["/opt/go/src/github.com/oysterbooks/halfshell/bin/halfshell"]

EXPOSE 8080
