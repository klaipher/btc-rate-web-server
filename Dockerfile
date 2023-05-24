FROM golang:1.20-alpine as build-stage

WORKDIR WORKDIR /build

RUN apk --no-cache add ca-certificates

COPY go.mod .
COPY main.go .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /rate-web-server .


FROM scratch

# Copy ca-certs for app web access
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-stage /rate-web-server /rate-web-server

CMD ["/rate-web-server"]