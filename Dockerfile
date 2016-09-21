FROM golang

ADD . /go/src/github.com/johnny-morrice/sensephreak

RUN go install github.com/johnny-morrice/sensephreak

ENTRYPOINT /go/bin/sensephreak

EXPOSE 1000-65535
