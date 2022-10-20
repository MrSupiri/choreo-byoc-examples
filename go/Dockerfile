FROM golang:1.18.0 as build-env

RUN mkdir /app
WORKDIR /app

# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/app -buildvcs=false

FROM alpine

COPY --from=build-env /go/bin/app /go/bin/app
USER 10014

ENTRYPOINT ["/go/bin/app"]
