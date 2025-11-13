FROM debian:13 AS binary-chooser

ARG TARGETPLATFORM

COPY ./build /build

RUN case ${TARGETPLATFORM} in \
         "linux/amd64")  BINARY=/build/bin/ipv4-for-ipv6_only-http-proxy_linux-amd64  ;; \
         "linux/arm64")  BINARY=/build/bin/ipv4-for-ipv6_only-http-proxy_linux-arm64  ;; \
    esac \
   && cp ${BINARY} /usr/bin/ipv4-for-ipv6_only-http-proxy
RUN chmod +x /usr/bin/ipv4-for-ipv6_only-http-proxy

FROM gcr.io/distroless/static:latest

LABEL org.opencontainers.image.source="https://github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy"
LABEL org.opencontainers.image.description="ipv4-for-ipv6_only-http-proxy"

WORKDIR /

COPY --from=binary-chooser /usr/bin/ipv4-for-ipv6_only-http-proxy /usr/bin/ipv4-for-ipv6_only-http-proxy
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/usr/bin/ipv4-for-ipv6_only-http-proxy"]
