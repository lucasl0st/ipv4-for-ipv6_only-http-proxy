# IPv4 to IPv6 Proxy

Proxy your IPv6-only HTTP infrastructure transparently using a single IPv4 host.

## Overview

The IPv4 to IPv6 Proxy is a solution for scenarios where endpoints are configured with IPv6-only networking,
but there's a need to accommodate IPv4 clients that cannot directly communicate with IPv6 endpoints.
This proxy acts as a bridge, translating incoming IPv4 requests to IPv6 requests, forwarding them to the target endpoint,
and returning the responses back to the clients.
The proxy uses dynamic hostname-based routing with wildcard SSL/TLS certificates.

## Prerequisites

Before you begin, ensure you have the following prerequisites:

- docker / docker-compose
- A server which can host a webserver on IPv4

## Usage

## Configuration

The proxy can be configured using environment variables. Available configuration options are:

- `HTTP_PORT`: The HTTP port the proxy should listen on. Default: `80`.
- `HTTPS_PORT`: The HTTPS port the proxy should listen on. Default: `443`.
- `CERT_DIR`: The directory containing the SSL/TLS certificate files. The `cert-file` and `key-file` need to be stored together in subdirectories. Default: `/etc/letsencrypt/live/`.
- `CERT_FILE_NAME`: The wildcard SSL/TLS certificate file for dynamic hostname-based routing. Default: `fullchain.pem`.
- `KEY_FILE_NAME`: The private key file corresponding to the wildcard certificate. Default: `privkey.pem`.
- `ALLOWED_HOSTS` Regex pattern for allowed hosts. Default: `.*`.

## Obtaining a Wildcard Certificate

To enable dynamic hostname-based routing with wildcard SSL/TLS certificates, you can use [Certbot](https://certbot.eff.org/)
to obtain a wildcard certificate from Let's Encrypt. The following steps outline the process:

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

4. You can now use the obtained wildcard certificate by placing them into `CERT_DIR` and by setting `CERT_FILE_NAME` and `KEY_FILE_NAME` accordingly (see [Configuration](#configuration)).
   Remember to renew the certificate before it expires. You can set up a cronjob to automatically renew the certificate using the following command:

   ```sh
   sudo certbot renew
   ```
For more detailed instructions and troubleshooting, refer to the [Certbot documentation](https://certbot.eff.org/docs/intro.html).

## Docker

Run use the [docker-compose.yaml](docker-compose.yaml) and run `docker-compose up -d` to start the latest version of the proxy.

### Configuring Docker for IPv6

To facilitate communication between containers and IPv6-only services within an IPv6-only environment,
Docker needs to be configured for IPv6 networking. Here's how to configure Docker for IPv6:

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
> [!WARNING]  
> In some cases a full system reboot might be required.

The docker-compose file in this repository includes the additionally necessary configuration for IPv6 networking, when using anything else you are on your own.
