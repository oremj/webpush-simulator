FROM golang:1.8

WORKDIR /go/src/github.com/oremj/webpush-simulator

COPY . .

RUN go-wrapper download && go-wrapper install

CMD  ["go-wrapper", "run"]
