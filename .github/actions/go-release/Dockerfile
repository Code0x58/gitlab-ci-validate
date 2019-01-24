FROM golang:1.11
RUN go get -u github.com/mitchellh/gox \
    && go get -u github.com/tcnksm/ghr
COPY go-release /usr/local/bin/
CMD go-release
