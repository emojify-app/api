FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN mkdir /service
RUN mkdir /service/cache

COPY ./emojify-api /service/
COPY ./images /service/images/

WORKDIR /service

CMD /service/emojify-api
