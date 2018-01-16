FROM golang:1.9-alpine AS build
COPY . $GOPATH/src/app
WORKDIR $GOPATH/src/app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch
COPY --from=build /go/src/app/app /app
EXPOSE 8080
CMD ["/app"]
