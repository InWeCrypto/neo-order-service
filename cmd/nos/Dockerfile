FROM golang:1.9.2-stretch

LABEL maintainer="support@inwecrypto.com"

COPY . /go/src/github.com/inwecrypto/neo-order-service

RUN go install github.com/inwecrypto/neo-order-service/cmd/nos && rm -rf /go/src

VOLUME ["/etc/inwecrypto/order/neo"]

WORKDIR /etc/inwecrypto/order/neo

CMD ["/go/bin/nos","--conf","/etc/inwecrypto/order/neo/neo-order-service.json"]