FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/management
WORKDIR /go/src/management

COPY ./ /go/src/management
RUN go get -v -d
RUN CGO_ENABLED=0 go build -o management

FROM scratch

COPY --from=0 /go/src/management/management /management

ENTRYPOINT ["/management"]