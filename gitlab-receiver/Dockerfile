FROM golang:1.18.2-alpine as build

WORKDIR /go/src/app
ADD . /go/src/app
RUN CGO_ENABLED=0 go build -o /gitlab-receiver -ldflags="-s -w" *.go

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=build /gitlab-receiver /
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

#########################################################
# This needs to be removed when deploying to Kubernetes
COPY direktiv.io* /
COPY config.yaml /
CMD ["/gitlab-receiver", "/config.yaml"]
########################################################

# CMD ["/gitlab-receiver", "/config/config.yaml"]

