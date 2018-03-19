FROM golang:1.9-alpine AS build
ARG VERSION=1.0.0
RUN apk --no-cache add git && \
  go get -u github.com/golang/dep/cmd/dep
COPY . $GOPATH/src/app
WORKDIR $GOPATH/src/app
RUN dep ensure && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.version=${VERSION}" -o app .

FROM scratch
COPY --from=build /go/src/app/app /app
EXPOSE 8080
CMD ["/app"]
