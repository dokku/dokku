FROM python:3.11.0-buster

EXPOSE 3001/udp
EXPOSE  3000/tcp
EXPOSE 3003

COPY . /app

WORKDIR /app
