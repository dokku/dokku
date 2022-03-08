FROM phusion/baseimage:focal-1.1.0

CMD ["/sbin/my_init"]

ARG DOKKU_TAG=0.17.7
ARG DOKKU_GID=200
ARG DOKKU_UID=200
ARG DOKKU_HOSTNAME=dokku.invalid
ARG DOKKU_SKIP_KEY_FILE=true
ARG DOKKU_VHOST_ENABLE=false
ARG DOKKU_WEB_CONFIG=false

RUN addgroup --gid $DOKKU_GID dokku \
  && adduser --uid $DOKKU_UID --gid $DOKKU_GID --disabled-password --gecos "" "dokku"

COPY ./tests/dhparam.pem /tmp/dhparam.pem
COPY ./build/package/ /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# hadolint ignore=DL3005,DL3008
RUN echo "dokku dokku/hostname string $DOKKU_HOSTNAME" | debconf-set-selections \
  && echo "dokku dokku/skip_key_file boolean $DOKKU_SKIP_KEY_FILE" | debconf-set-selections \
  && echo "dokku dokku/vhost_enable boolean $DOKKU_VHOST_ENABLE" | debconf-set-selections \
  && curl -sSL https://packagecloud.io/dokku/dokku/gpgkey | apt-key add - \
  && echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ bionic main" | tee /etc/apt/sources.list.d/dokku.list \
  && mkdir -p /etc/nginx/ \
  && cp /tmp/dhparam.pem /etc/nginx/dhparam.pem \
  && apt-get update -qq \
  && apt-get upgrade -qq -y \
  && apt-get -qq -y --no-install-recommends --only-upgrade install openssl openssh-server \
  && apt-get -qq -y --no-install-recommends install rsync "/tmp/dokku-$(dpkg --print-architecture).deb" \
  && apt-get purge -qq -y syslog-ng-core \
  && apt-get autoremove -qq -y \
  && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /tmp

COPY ./docker .

RUN \
  rsync -a /tmp/ / \
  && rm -rf /tmp/* \
  && rm /etc/runit/runsvdir/default/sshd/down \
  && chown -R dokku:dokku /home/dokku/ \
  && mkdir -p /skel/etc /skel/home /skel/var/lib/dokku /var/log/services \
  && mv /etc/ssh /skel/etc/ssh \
  && mv /home/dokku /skel/home/dokku \
  && mv /var/lib/dokku/config /skel/var/lib/dokku/config \
  && mv /var/lib/dokku/data /skel/var/lib/dokku/data \
  && ln -sf /mnt/dokku/etc/ssh /etc/ssh \
  && ln -sf /mnt/dokku/home/dokku /home/dokku \
  && ln -sf /mnt/dokku/var/lib/dokku/config /var/lib/dokku/config \
  && ln -sf /mnt/dokku/var/lib/dokku/data /var/lib/dokku/data \
  && mv /etc/my_init.d/00_regen_ssh_host_keys.sh /etc/my_init.d/15_regen_ssh_host_keys \
  && rm -f /etc/nginx/sites-enabled/default /usr/share/nginx/html/index.html /etc/my_init.d/10_syslog-ng.init \
  && rm -f /usr/local/openresty/nginx/conf/sites-enabled/default /usr/share/openresty/html/index.html \
  && sed -i '/imklog/d' /etc/rsyslog.conf \
  && rm -f /var/log/btmp /var/log/wtmp /var/log/*log /var/log/apt/* /var/log/dokku/*.log /var/log/nginx/* /var/log/openresty/*
