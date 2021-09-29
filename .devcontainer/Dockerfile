FROM dokku/dokku:latest

RUN apt-get update
RUN apt-get install --no-install-recommends -y build-essential file nano && \
  apt-get install --no-install-recommends -y help2man shellcheck uuid-runtime wget xmlstarlet && \
  apt-get clean autoclean && \
  apt-get autoremove --yes && \
  rm -rf /var/lib/apt/lists/*

ADD https://raw.githubusercontent.com/dokku/dokku/master/tests/dhparam.pem /mnt/dokku/etc/nginx/dhparam.pem

COPY .devcontainer/bin/ /usr/local/bin/
COPY ["tests.mk", "Makefile"]
RUN make ci-dependencies

COPY . .

ENV DOKKU_HOSTNAME=dokku.me
LABEL org.label-schema.schema-version=1.0 org.label-schema.vendor=dokku com.dokku.devcontainer=true
