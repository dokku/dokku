FROM python:3.11.0-buster

ARG BUILD_ARG=key

WORKDIR /app

COPY . /app

RUN echo $BUILD_ARG > BUILD_ARG.contents
