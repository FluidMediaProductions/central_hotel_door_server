FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/hotels
WORKDIR /go/src/hotels

COPY ./ /go/src/hotels
RUN go get -v -d
RUN CGO_ENABLED=0 go build -o hotels

FROM scratch

COPY --from=0 /go/src/hotels/hotels /hotels

ENTRYPOINT ["/hotels"]