FROM ubuntu:14.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update

RUN apt-get -y install gcc git build-essential wget ruby-dev ruby1.9.1 lintian rpm help2man man-db

RUN command -v fpm > /dev/null || sudo gem install fpm --no-ri --no-rdoc

WORKDIR /dokku

COPY Makefile /dokku/

COPY *.mk /dokku/

RUN make deb-setup rpm-setup

COPY . /dokku

RUN make sshcommand plugn version copyfiles

ARG DOKKU_VERSION=master
ENV DOKKU_VERSION ${DOKKU_VERSION}

ARG DOKKU_GIT_REV
ENV DOKKU_GIT_REV ${DOKKU_GIT_REV}

ARG IS_RELEASE=false
ENV IS_RELEASE ${IS_RELEASE}

RUN make deb-herokuish deb-dokku deb-plugn deb-sshcommand deb-sigil

RUN mkdir -p /data && cp /tmp/*.deb /data && ls /data/
