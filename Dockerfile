FROM ubuntu
COPY ./build/server /bin/broker
CMD /bin/broker
