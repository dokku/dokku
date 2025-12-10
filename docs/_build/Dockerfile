FROM python:3.15.0a2-alpine

WORKDIR /usr/src/app

COPY docs/_build/requirements.txt .

RUN apk add --no-cache bash build-base yaml && \
  pip install --no-cache-dir -r requirements.txt

COPY docs/_build/entrypoint /bin/entrypoint

ENTRYPOINT [ "/bin/entrypoint" ]
