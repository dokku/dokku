FROM python:3.12.6-bookworm

EXPOSE 3001/udp
EXPOSE  3000/tcp
EXPOSE 3003

COPY . /app

WORKDIR /app

ENTRYPOINT ["/app/exec-entrypoint.sh"]
