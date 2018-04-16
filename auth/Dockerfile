FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/auth
WORKDIR /go/src/auth

RUN go get -v -u github.com/golang/dep/...
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v --vendor-only

COPY ./ /go/src/auth
RUN CGO_ENABLED=0 go build -o auth

FROM scratch

COPY --from=0 /go/src/auth/auth /auth

ENTRYPOINT ["/auth"]