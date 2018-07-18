FROM alpine

RUN mkdir /service

COPY ./emojifyServer /service/
COPY ./images /service/images/

WORKDIR /service

CMD /service/emojifyServer
