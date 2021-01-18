FROM dokku/build-base:0.0.1

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -qq && apt-get -qq -y --no-install-recommends install gcc git build-essential wget ruby-dev ruby1.9.1 lintian rpm help2man man-db
RUN command -v fpm >/dev/null || sudo gem install fpm --no-ri --no-rdoc

ARG WORKDIR=/go/src/github.com/dokku/dokku

WORKDIR ${WORKDIR}

COPY . ${WORKDIR}

RUN make deb-dokku-update rpm-dokku-update

RUN ls -lha /tmp/
