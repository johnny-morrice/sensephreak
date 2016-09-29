FROM golang

ADD . /go/src/github.com/johnny-morrice/sensephreak

RUN go install github.com/johnny-morrice/sensephreak

ENTRYPOINT ["/go/src/github.com/johnny-morrice/sensephreak/sensephreak-docker.sh"]

EXPOSE 1-65535
