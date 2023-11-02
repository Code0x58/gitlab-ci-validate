FROM golang:alpine as build-step

WORKDIR /build
COPY . .

RUN apk add --no-cache ca-certificates && \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' ./gitlab-ci-validate.go


FROM scratch

WORKDIR /yaml
COPY --from=build-step /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-step /build/gitlab-ci-validate /gitlab-ci-validate

ENTRYPOINT [ "/gitlab-ci-validate" ]
CMD [ ".gitlab-ci.yml" ]
