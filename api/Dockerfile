FROM golang:alpine

RUN apk add --update --no-cache git mercurial

RUN go get -u github.com/gorilla/mux \
    && go get -u github.com/gorilla/handlers \
    && go get -u github.com/rs/cors \
    && go get -u github.com/go-sql-driver/mysql \
    && go get -u golang.org/x/image/draw \
    && go get -u github.com/disintegration/imaging \
    && apk del git mercurial
WORKDIR /go/src/api
COPY . .
RUN go install -v ./...
RUN mkdir -p /bricks
VOLUME ["/bricks"]
CMD ["api"]