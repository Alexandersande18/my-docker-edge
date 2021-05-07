FROM golang:1.16
ADD . $GOPATH/src
WORKDIR $GOPATH/src/src
RUN go build main.go
EXPOSE 11451
ENTRYPOINT ["./main"]