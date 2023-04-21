FROM node:20-alpine

RUN apk add --no-cache bash

EXPOSE 3001/udp
EXPOSE  3000/tcp
EXPOSE 3003

COPY . /app
WORKDIR /app
RUN npm install

CMD npm start
