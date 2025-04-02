FROM golang:1.24-alpine as builder

RUN apk add git

WORKDIR /src

COPY go.mod ./
COPY go.sum ./

RUN mkdir /new_tmp

RUN go mod download

COPY . ./

RUN go build -o /ipv4-for-ipv6_only-http-proxy

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy"
LABEL org.opencontainers.image.description="ipv4-for-ipv6_only-http-proxy"

COPY --from=builder /ipv4-for-ipv6_only-http-proxy /usr/bin/ipv4-for-ipv6_only-http-proxy

COPY --from=builder /new_tmp /tmp
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 80
EXPOSE 443

CMD [ "/usr/bin/ipv4-for-ipv6_only-http-proxy" ]