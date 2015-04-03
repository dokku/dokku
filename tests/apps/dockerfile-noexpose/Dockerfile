FROM ubuntu:trusty
ENV LC_ALL C
ENV DEBIAN_FRONTEND noninteractive
ENV DEBCONF_NONINTERACTIVE_SEEN true

RUN apt-get install -y software-properties-common && add-apt-repository ppa:chris-lea/node.js && apt-get update
RUN apt-get install -y build-essential curl postgresql-client-9.3 nodejs git

COPY . /app
WORKDIR /app
RUN npm install

CMD npm start
