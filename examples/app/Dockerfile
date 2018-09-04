FROM golang:1.11 AS build
ENV GOPATH=""
ENV GO111MODULE=on
ARG LDFLAGS
COPY . /go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "$LDFLAGS" -o app .

FROM scratch
COPY --from=build /go/app /app
EXPOSE 8080
CMD ["/app"]
