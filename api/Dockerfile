FROM golang:alpine

RUN apk add --update --no-cache gcc libc-dev mesa-dev libxrandr-dev libxcursor-dev libxi-dev libxinerama-dev alpine-sdk  cmake musl-dev git mercurial

RUN mkdir /home/git; \
    cd /home/git; \
    sudo git clone https://github.com/assimp/assimp.git -b master; 

RUN cd /home/git/assimp; \
    cmake CMakeLists.txt -G 'Unix Makefiles'; \
    make; \
    make install; 
    #ldconfig;

 RUN assimp help

RUN go get -u github.com/gorilla/mux \
    && go get -u github.com/gorilla/handlers \
    && go get -u github.com/rs/cors \
    && go get -u github.com/go-sql-driver/mysql \
    && go get -u golang.org/x/image/draw \
    && go get -u github.com/go-gl/gl/v4.1-core/gl \
    && go get -u github.com/raedatoui/glutils \
    && apk del git mercurial
WORKDIR /go/src/api
COPY . .
RUN go install -v ./...
RUN mkdir -p /bricks
VOLUME ["/bricks"]
CMD ["api"]