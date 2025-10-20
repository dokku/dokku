FROM ruby:3.4.7

RUN mkdir /app
WORKDIR /app

ADD Procfile Procfile
ADD exec-entrypoint exec-entrypoint
ADD app.json app.json

ENTRYPOINT ["./exec-entrypoint"]
