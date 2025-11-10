FROM phusion/baseimage:noble-1.0.2

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

ENV DOKKU_INIT_SYSTEM=sv

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# hadolint ignore=DL3005,DL3008
RUN mkdir -p /etc/apt/keyrings \
  && mkdir -p /etc/apt/keyrings \
  && apt-get remove -y systemd && apt-get autoremove -y && apt-get update && apt-get -y --no-install-recommends install gpg lsb-release openssl openssh-server rsync software-properties-common \
  && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
  && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
  && add-apt-repository ppa:cncf-buildpacks/pack-cli \
  && apt-get update \
  && apt-get -y --no-install-recommends install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin pack-cli \
  && curl -o /tmp/nixpacks.bash -sSL https://nixpacks.com/install.sh \
  && chmod +x /tmp/nixpacks.bash \
  && NIXPACKS_BIN_DIR=/usr/bin BIN_DIR=/usr/bin /tmp/nixpacks.bash \
  && test -x /usr/bin/nixpacks \
  && rm -rf /tmp/nixpacks.bash \
  && curl -o /tmp/railpack.bash -sSL https://railpack.com/install.sh \
  && chmod +x /tmp/railpack.bash \
  && RAILPACK_BIN_DIR=/usr/bin BIN_DIR=/usr/bin /tmp/railpack.bash \
  && test -x /usr/bin/railpack \
  && rm -rf /tmp/railpack.bash \
  && echo "dokku dokku/hostname string $DOKKU_HOSTNAME" | debconf-set-selections \
  && echo "dokku dokku/skip_key_file boolean $DOKKU_SKIP_KEY_FILE" | debconf-set-selections \
  && echo "dokku dokku/vhost_enable boolean $DOKKU_VHOST_ENABLE" | debconf-set-selections \
  && curl -sSL https://packagecloud.io/dokku/dokku/gpgkey | apt-key add - \
  && echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ noble main" | tee /etc/apt/sources.list.d/dokku.list \
  && mkdir -p /etc/nginx/ \
  && cp /tmp/dhparam.pem /etc/nginx/dhparam.pem \
  && apt-get update \
  && apt-get upgrade -y \
  && apt-get -y --no-install-recommends install lambda-builder "/tmp/dokku-$(dpkg --print-architecture).deb" \
  && apt-get purge -y syslog-ng-core \
  && apt-get autoremove -y \
  && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /tmp

COPY ./docker .

RUN \
  rsync -a /tmp/ / \
  && chmod +x /usr/local/bin/* \
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
  && ln -sf /mnt/dokku/var/lib/dokku/services /var/lib/dokku/services \
  && mv /etc/my_init.d/00_regen_ssh_host_keys.sh /etc/my_init.d/15_regen_ssh_host_keys \
  && rm -f /etc/nginx/sites-enabled/default /usr/share/nginx/html/index.html /etc/my_init.d/10_syslog-ng.init \
  && rm -f /usr/local/openresty/nginx/conf/sites-enabled/default /usr/share/openresty/html/index.html \
  && sed -i '/imklog/d' /etc/rsyslog.conf \
  && rm -f /var/log/btmp /var/log/wtmp /var/log/*log /var/log/apt/* /var/log/dokku/*.log /var/log/nginx/* /var/log/openresty/*
