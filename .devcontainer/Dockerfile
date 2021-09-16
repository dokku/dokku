FROM dokku/dokku:latest

RUN apt-get update
RUN apt-get install --no-install-recommends -y build-essential file nano && \
  apt-get install --no-install-recommends -y shellcheck uuid-runtime xmlstarlet && \
  apt-get clean autoclean && \
  apt-get autoremove --yes && \
  rm -rf /var/lib/apt/lists/*

ADD https://raw.githubusercontent.com/dokku/dokku/master/tests/dhparam.pem /mnt/dokku/etc/nginx/dhparam.pem

COPY .devcontainer/bin/ /usr/local/bin/
COPY . .

RUN make ci-dependencies

ENV DOKKU_HOSTNAME=dokku.me
