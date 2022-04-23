FROM dokku/dokku:latest

RUN apt-get update
RUN apt-get install --no-install-recommends -y build-essential file nano && \
  apt-get install --no-install-recommends -y help2man shellcheck uuid-runtime wget xmlstarlet && \
  apt-get clean autoclean && \
  apt-get autoremove --yes && \
  rm -rf /var/lib/apt/lists/*

RUN wget https://dl.google.com/go/go1.17.9.linux-amd64.tar.gz && \
  tar -xvf go1.17.9.linux-amd64.tar.gz && \
  mv go /usr/local

RUN GOROOT=/usr/local/go /usr/local/go/bin/go install golang.org/x/tools/gopls@latest 2>&1

ADD https://raw.githubusercontent.com/dokku/dokku/master/tests/dhparam.pem /mnt/dokku/etc/nginx/dhparam.pem

COPY .devcontainer/bin/ /usr/local/bin/
COPY ["tests.mk", "Makefile"]
RUN make ci-dependencies

COPY . .

ENV DOKKU_HOSTNAME=dokku.me GOROOT=/usr/local/go PATH=/usr/local/go/bin:/root/go/bin:$PATH PLUGIN_MAKE_TARGET=build

LABEL org.label-schema.schema-version=1.0 org.label-schema.vendor=dokku com.dokku.devcontainer=true
