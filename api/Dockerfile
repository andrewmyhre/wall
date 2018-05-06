FROM golang:alpine

RUN apk add --no-cache git mercurial \
    && go get -u github.com/gorilla/mux \
    && go get -u github.com/gorilla/handlers \
    && go get -u github.com/rs/cors \
    && go get -u github.com/go-sql-driver/mysql \
    && apk del git mercurial
WORKDIR /go/src/api
COPY . .
RUN go install -v ./...
RUN mkdir /bricks
CMD ["api"]