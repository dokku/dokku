FROM ruby:4.0.8

RUN mkdir /app
WORKDIR /app

ADD Procfile Procfile
ADD entrypoint entrypoint
ADD app.json app.json

ENTRYPOINT ["./entrypoint"]
