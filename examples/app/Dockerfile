FROM golang:1.9-alpine AS build
ARG VERSION=1.0.0
COPY . $GOPATH/src/app
WORKDIR $GOPATH/src/app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.version=${VERSION}" -o app .

FROM scratch
COPY --from=build /go/src/app/app /app
EXPOSE 8080
CMD ["/app"]
