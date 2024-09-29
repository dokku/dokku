FROM python:3.12.6-bookworm

ARG BUILD_ARG=key

WORKDIR /app

COPY . /app

RUN echo $BUILD_ARG > BUILD_ARG.contents
