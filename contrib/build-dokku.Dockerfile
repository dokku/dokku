FROM dokku/build-base:0.0.1 AS builder

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -qq && apt-get -qq -y --no-install-recommends install gcc git build-essential wget ruby-dev ruby1.9.1 lintian rpm help2man man-db
RUN command -v fpm >/dev/null || sudo gem install fpm --no-ri --no-rdoc

ARG GOLANG_VERSION

RUN wget -qO /tmp/go${GOLANG_VERSION}.linux-amd64.tar.gz https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz \
  && tar -C /usr/local -xzf /tmp/go${GOLANG_VERSION}.linux-amd64.tar.gz \
  && cp /usr/local/go/bin/* /usr/local/bin

ARG WORKDIR=/go/src/github.com/dokku/dokku

WORKDIR ${WORKDIR}

COPY Makefile ${WORKDIR}/
COPY *.mk ${WORKDIR}/

RUN make deb-setup sshcommand plugn

COPY . ${WORKDIR}

ENV GOPATH=/go

FROM builder as amd64

ARG PLUGIN_MAKE_TARGET
ARG DOKKU_VERSION=master
ARG DOKKU_GIT_REV
ARG IS_RELEASE=false

RUN PLUGIN_MAKE_TARGET=${PLUGIN_MAKE_TARGET} \
  DOKKU_VERSION=${DOKKU_VERSION} \
  DOKKU_GIT_REV=${DOKKU_GIT_REV} \
  IS_RELEASE=${IS_RELEASE} \
  SKIP_GO_CLEAN=true \
  make version copyfiles \
  && make deb-dokku

FROM builder as armhf

COPY --from=amd64 /tmp /tmp
COPY --from=amd64 /usr/local/share/man/man1/dokku.1 /usr/local/share/man/man1/dokku.1-generated

RUN rm -rf /tmp/build-dokku

ARG PLUGIN_MAKE_TARGET
ARG DOKKU_VERSION=master
ARG DOKKU_GIT_REV
ARG IS_RELEASE=false

RUN PLUGIN_MAKE_TARGET=${PLUGIN_MAKE_TARGET} \
  DOKKU_VERSION=${DOKKU_VERSION} \
  DOKKU_GIT_REV=${DOKKU_GIT_REV} \
  IS_RELEASE=${IS_RELEASE} \
  SKIP_GO_CLEAN=true \
  GOARCH=arm make version copyfiles \
  && DOKKU_ARCHITECTURE=armhf GOARCH=arm make deb-dokku

FROM builder as arm64

COPY --from=armhf /tmp /tmp
COPY --from=amd64 /usr/local/share/man/man1/dokku.1 /usr/local/share/man/man1/dokku.1-generated

RUN rm -rf /tmp/build-dokku

ARG PLUGIN_MAKE_TARGET
ARG DOKKU_VERSION=master
ARG DOKKU_GIT_REV
ARG IS_RELEASE=false

RUN PLUGIN_MAKE_TARGET=${PLUGIN_MAKE_TARGET} \
  DOKKU_VERSION=${DOKKU_VERSION} \
  DOKKU_GIT_REV=${DOKKU_GIT_REV} \
  IS_RELEASE=${IS_RELEASE} \
  SKIP_GO_CLEAN=true \
  GOARCH=arm64 make version copyfiles \
  && DOKKU_ARCHITECTURE=arm64 GOARCH=arm64 make deb-dokku

RUN ls -lha /tmp/
