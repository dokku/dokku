FROM phusion/baseimage:0.11

CMD ["/sbin/my_init"]

RUN apt-get update && apt-get -y upgrade && apt-get -y install \
    nano \
    rsync \
    software-properties-common \
    sudo \
    wget && \
    apt-get purge -y syslog-ng-core && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ARG DOKKU_TAG=0.17.7
ARG DOKKU_GID=200
ARG DOKKU_UID=200
ARG DOKKU_DOCKERFILE=true
ARG DOKKU_WEB_CONFIG=false
ARG DOKKU_HOSTNAME=dokku.invalid
ARG DOKKU_VHOST_ENABLE=false
ARG DOKKU_SKIP_KEY_FILE=true

RUN addgroup --gid $DOKKU_GID dokku && \
    adduser --uid $DOKKU_UID --gid $DOKKU_GID --disabled-password --gecos "" "dokku"

RUN mkdir /etc/nginx/
COPY tests/dhparam.pem /etc/nginx/dhparam.pem

RUN export DOKKU_TAG=$DOKKU_TAG \
      DOKKU_DOCKERFILE=$DOKKU_DOCKERFILE \
      DOKKU_WEB_CONFIG=$DOKKU_WEB_CONFIG \
      DOKKU_HOSTNAME=$DOKKU_HOSTNAME \
      DOKKU_VHOST_ENABLE=$DOKKU_VHOST_ENABLE \
      DOKKU_SKIP_KEY_FILE=$DOKKU_SKIP_KEY_FILE && \
    wget https://raw.githubusercontent.com/dokku/dokku/v$DOKKU_TAG/bootstrap.sh && \
    bash bootstrap.sh && \
    apt-get clean && rm -rf bootstrap.sh /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /tmp

COPY ./docker .

RUN \
    rsync -a /tmp/ / && \
    rm -rf /tmp/* && \
    rm /etc/runit/runsvdir/default/sshd/down && \
    chown -R dokku:dokku /home/dokku/ && \
    mkdir -p /skel/etc /skel/home /skel/var/lib/dokku /var/log/services && \
    mv /etc/ssh /skel/etc/ssh && \
    mv /home/dokku /skel/home/dokku && \
    mv /var/lib/dokku/config /skel/var/lib/dokku/config && \
    mv /var/lib/dokku/data /skel/var/lib/dokku/data && \
    ln -sf /mount/etc/ssh /etc/ssh && \
    ln -sf /mount/home/dokku /home/dokku && \
    ln -sf /mount/var/lib/dokku/config /var/lib/dokku/config && \
    ln -sf /mount/var/lib/dokku/data /var/lib/dokku/data && \
    mv /etc/my_init.d/00_regen_ssh_host_keys.sh \
      /etc/my_init.d/15_regen_ssh_host_keys && \
    rm /etc/nginx/sites-enabled/default && \
    rm /usr/share/nginx/html/index.html && \
    rm /etc/my_init.d/10_syslog-ng.init && \
    sed -i '/imklog/d' /etc/rsyslog.conf && \
    rm /var/log/btmp /var/log/wtmp /var/log/*log /var/log/apt/* /var/log/dokku/* /var/log/nginx/*
