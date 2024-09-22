FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -qq && apt-get -qq -y --no-install-recommends install \
    adduser \
    build-essential \
    ca-certificates \
    coreutils \
    curl \
    gcc \
    git \
    help2man \
    jq \
    libc-bin \
    lintian \
    man-db \
    openssh-client \
    python3 \
    rpm \
    ruby-dev \
    sudo \
    wget
RUN command -v fpm || gem install fpm
