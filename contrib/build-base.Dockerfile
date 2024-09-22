FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -qq && apt-get -qq -y --no-install-recommends install ca-certificates curl gcc git jq build-essential wget ruby-dev lintian python3 rpm help2man man-db sudo
RUN command -v fpm || gem install fpm
