FROM golang:1.9.2-stretch

LABEL maintainer="support@inwecrypto.com"

COPY . /go/src/github.com/inwecrypto/neo-order-service

RUN go install github.com/inwecrypto/neo-order-service/cmd/neo-orders && rm -rf /go/src

VOLUME ["/etc/inwecrypto/order/neo"]

WORKDIR /etc/inwecrypto/order/neo

EXPOSE 8000

CMD ["/go/bin/neo-orders","--conf","/etc/inwecrypto/order/neo/neo-order-service.json"]