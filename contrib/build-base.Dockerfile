FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -qq && apt-get -qq -y --no-install-recommends install ca-certificates curl gcc git jq build-essential wget ruby-dev lintian python3 rpm help2man man-db sudo
RUN curl -sL -o /usr/local/share/ca-certificates/GlobalSignRootCA_R3.crt https://raw.githubusercontent.com/rubygems/rubygems/master/lib/rubygems/ssl_certs/rubygems.org/GlobalSignRootCA_R3.pem
RUN update-ca-certificates
RUN command -v fpm || gem install fpm
