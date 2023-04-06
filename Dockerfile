FROM golang:alpine3.16 as builder

RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init 
RUN  apk add librdkafka-dev pkgconf
RUN apk add make
#
RUN mkdir -p $GOPATH/src/gitlab.7i.uz/invan/invan_catalog_service 
WORKDIR $GOPATH/src/gitlab.7i.uz/invan/invan_catalog_service

# Copy the local package files to the container's workspace.
COPY . ./


# installing depends and build
RUN export CGO_ENABLED=1 && \
  export GOOS=linux && \
  make build && \
  mv ./bin/invan_catalog_service /


FROM alpine
RUN  apk add librdkafka-dev pkgconf
COPY --from=builder invan_catalog_service .
ENTRYPOINT ["/invan_catalog_service"]
