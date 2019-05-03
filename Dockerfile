FROM ubuntu:latest

RUN apt-get update && apt-get install -y ca-certificates

RUN mkdir /service
RUN mkdir /service/cache

COPY ./emojify-api /service/
COPY ./images /service/images/

WORKDIR /service

ENTRYPOINT /service/emojify-api
