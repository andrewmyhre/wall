FROM golang:alpine

RUN apk add --update --no-cache gcc libc-dev mesa-dev libxrandr-dev libxcursor-dev libxi-dev libxinerama-dev alpine-sdk  cmake musl-dev git mercurial

RUN go get std \
	&& go get -u github.com/go-gl/gl/v4.1-core/gl \
	&& go get -u github.com/go-gl/glfw/v3.1/glfw
WORKDIR /go/src/treater
COPY . .
RUN go install -v ./...