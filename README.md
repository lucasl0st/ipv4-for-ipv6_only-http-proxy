# IPv4 to IPv6 Proxy

This project provides a Go-based proxy server that acts as an intermediary between IPv4 clients and IPv6-only endpoints. It allows IPv4 clients to communicate with services deployed on IPv6-only infrastructure seamlessly. The proxy supports SSL/TLS termination using wildcard certificates for dynamic hostname-based routing.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
    - [Build](#build)
    - [Run](#run)
- [Configuration](#configuration)
- [Obtaining a Wildcard Certificate](#obtaining-a-wildcard-certificate)
- [Dynamic Routing](#dynamic-routing)
- [Dockerization](#dockerization)
- [Configuring DNS ENtries using external-dns](#Configuring-DNS-Entries-using-external-dns)

## Overview

The IPv4 to IPv6 Proxy is a solution for scenarios where endpoints are configured with IPv6-only networking, but there's a need to accommodate IPv4 clients that cannot directly communicate with IPv6 endpoints. This proxy acts as a bridge, translating incoming IPv4 requests to IPv6 requests, forwarding them to the target endpoint, and returning the responses back to the clients. The proxy also supports dynamic hostname-based routing using wildcard SSL/TLS certificates.

## Prerequisites

Before you begin, ensure you have the following prerequisites:

- Go programming environment (for building the proxy)
- Docker (for containerization)
- `docker-compose` (for managing the deployment)
- A working network infrastructure with IPv6-only endpoints

## Usage

### Build

To build the proxy, follow these steps:

Clone the repository:

```sh
git clone https://github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy.git
```

Change to the project directory by running:

```sh
cd ipv4-for-ipv6_only-http-proxy
```


Build the proxy binary by running:

```sh
go build -o ipv4-to-ipv6-proxy
```

### Run

To run the proxy, execute the following command:

```sh
./ipv4-to-ipv6-proxy
```


The proxy will start and listen for incoming connections on the specified port (default is 8080). Make sure to set the necessary environment variables for configuring ports and SSL/TLS certificates (see [Configuration](#configuration)).

## Configuration

The proxy can be configured using environment variables. Available configuration options are:

- `HTTP_PORT`: The HTTP port the proxy should listen on. Default: `80`.
- `HTTPS_PORT`: The HTTPS port the proxy should listen on. Default: `443`.
- `CERT_FILE`: The wildcard SSL/TLS certificate file for dynamic hostname-based routing. Default: `cert.pem`.
- `KEY_FILE`: The private key file corresponding to the wildcard certificate. Default: `key.pem`.
- `ALLOWED_HOSTS` Regex pattern for allowed hosts. Default: `.*`.

You can set these environment variables before running the proxy.

## Obtaining a Wildcard Certificate

To enable dynamic hostname-based routing with wildcard SSL/TLS certificates, you can use [Certbot](https://certbot.eff.org/) to obtain a wildcard certificate from Let's Encrypt. The following steps outline the process:

1. Install Certbot on your system if it's not already installed:

   Run:

   ```sh
   sudo apt-get install certbot
   ```

2. Obtain a wildcard certificate using the DNS-01 challenge method. Replace `YOUR_EMAIL` with your email address and `YOUR_DOMAIN` with your actual domain:

   Run:

   ```sh
   sudo certbot certonly --manual --preferred-challenges=dns --email YOUR_EMAIL --server https://acme-v02.api.letsencrypt.org/directory --agree-tos -d *.YOUR_DOMAIN
   ```
   This command will prompt you to add a DNS TXT record to your domain's DNS settings to prove domain ownership. Follow the instructions provided by Certbot.

3. Once you've added the DNS TXT record and the domain ownership is verified, Certbot will issue a wildcard certificate and save it in the default Let's Encrypt path.

4. You can now use the obtained wildcard certificate and corresponding private key for the `CERT_FILE` and `KEY_FILE` environment variables when running the proxy (see [Configuration](#configuration)).
   Remember to renew the certificate before it expires. You can set up a cron job to automatically renew the certificate using the following command:

   ```sh
   sudo certbot renew
   ```


For more detailed instructions and troubleshooting, refer to the [Certbot documentation](https://certbot.eff.org/docs/intro.html).

## Dynamic Routing

The proxy supports dynamic hostname-based routing using wildcard SSL/TLS certificates. When a request is received, the proxy extracts the hostname, looks up the corresponding IPv6 address in your infrastructure, and forwards the request to that address. This allows seamless communication between IPv4 clients and IPv6-only services.

## Dockerization

A Docker image can be built

```sh
docker build -t ipv4-to-ipv6-proxy .
```

You can then use the included `docker-compose.yaml` file to deploy the proxy alongside your infrastructure.

### Configuring Docker for IPv6

To facilitate communication between containers and IPv6-only services within an IPv6-only environment, Docker needs to be configured for IPv6 networking. Here's how to configure Docker for IPv6:

#### Why IPv6 Configuration?

In scenarios where IPv6 is the primary networking protocol, Docker containers must use IPv6 addresses for communication. By default, Docker is configured for IPv4 networking. To enable containers to communicate via IPv6, Docker needs specific configuration.

#### Configuration Steps

1. Open the Docker daemon configuration file in a text editor. On Linux, use the following command to open the file:

    ```sh
    sudo nano /etc/docker/daemon.json
    ```

2. Add the following configuration lines to enable IPv6 support and define an IPv6 CIDR block:

    ```json
    {
      "ipv6": true,
      "fixed-cidr-v6": "fd4c:f221::/64",
      "experimental": true,
      "ip6tables": true,
      "default-address-pools": [
        { "base": "172.17.0.0/16", "size": 16 },
        { "base": "172.18.0.0/16", "size": 16 },
        { "base": "172.19.0.0/16", "size": 16 },
        { "base": "172.20.0.0/14", "size": 16 },
        { "base": "172.24.0.0/14", "size": 16 },
        { "base": "172.28.0.0/14", "size": 16 },
        { "base": "192.168.0.0/16", "size": 20 },
        { "base": "fd4c:f222::/104", "size": 112}
      ]
    }
    ```
   
3. Restart the Docker daemon to apply the changes
    
    ```sh
    sudo systemctl restart docker
    ```

The docker-compose file in this repository includes the additionally necessary configuration for IPv6 networking, when using anything else you are on your own.

## Configuring DNS Entries using external-dns

If you want to configure the DNS IPv4 entries using [external-dns](https://github.com/kubernetes-sigs/external-dns), you can use a minimalistic Kubernetes Ingress resource to trigger external-dns.

Here's a simple example of how to configure a "fake" Ingress resource to trigger external-dns to add DNS records:

Create an Ingress resource with the `external-dns.alpha.kubernetes.io/hostname` and `external-dns.alpha.kubernetes.io/target` annotations:


```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: sub-example-ipv4
  annotations:
    external-dns.alpha.kubernetes.io/hostname: sub.example.com
    external-dns.alpha.kubernetes.io/target: 1.2.3.4
spec:
  rules:
  - http:
```