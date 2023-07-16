FROM golang:1.20 AS builder

WORKDIR $GOPATH/src/caster

COPY . ./

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /caster main.go

FROM ubuntu:23.10

WORKDIR /

RUN apt-get -y update && apt-get -y install curl

COPY --from=builder /caster ./

CMD ["./caster", "run"]
