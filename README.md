Proxy for OCSP stapling
=======================

This repository already build as a handy docker image https://hub.docker.com/r/cooolin/ocsp-proxy.

OCSP stapling means that the SSL server (rather than client) has to make requests to CA servers
for revoked certificates lists, making the check faster and more reliable for clients.

If:

* you're not allowed to connect from your SSL servers to the CA server because of a firewall,
* and your SSL server allows you to force the URL of the OCSP server
* but not of a HTTP proxy

then this tool may help you.

# Usage

## General usage

First get the OCSP host of your certificate:

```sh
openssl x509 -in certificate.crt -noout -text | grep OCSP
# eg. OCSP - URI:http://ocsp.int-x3.letsencrypt.org
```

Then start the proxy with `ocsphost` (required) and `http` (optinal, default listen on port 8080):

```sh
docker run -d --rm --name ocsp-proxy -p 8080:8080 \
           -e HTTP_PROXY=http://YOUR_PROXY:8888 \
           -e ocsphost='http://ocsp.int-x3.letsencrypt.org' \
           -e http=':8080' \
           cooolin/ocsp-proxy
```

it will listen on port 8080 for HTTP request and will forward the request to the `ocsphost`.

Then configure your nginx:

```nginx
ssl_stapling on;
ssl_stapling_verify on;
ssl_stapling_responder http://127.0.0.1:8080/; 
ssl_trusted_certificate /etc/ssl/ca-certs.pem;  # as the same as `ssl_certificate`
```

> NOTE:
> we use `127.0.0.1` by assuming nginx is not in a docker container.

## Docker compose usage

For `docker-compose` usage, we recommend configure `nginx` and `ocsp-proxy` in an united `docker-compose.yml`, which can easily visit each other by container name rather than IP address.

Add OCSP proxy service to `docker-compose.yml`:

```yaml
version: '3.4'

services:

  # nginx service
  # ...

  ocsp-proxy:
    image: cooolin/ocsp-proxy
    container_name: ocsp-proxy
    environment:
      - HTTP_PROXY=http://YOUR_PROXY:8888
      - ocsphost=http://ocsp.int-x3.letsencrypt.org
      - http=:8080
    restart: always
```

in `nginx.conf`:

```nginx
server {
    listen 443 ssl;

    # ssl basic configures
    # ...

    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_stapling_responder http://ocsp-proxy:8080/;  # `ocsp-proxy` is the `container_name` of OSCP proxy service
    ssl_trusted_certificate /certs/certificate.crt;  # as the same as ssl_certificate

    # ...
}
```

# What if my server is outside of firewall?

No OCSP proxy needed, just open OCSP stapling.

in `nginx.conf`:

```nginx
http {
    # ...

    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 10s;

    server {
        # ...
        ssl_stapling on;
        ssl_stapling_verify on;
        ssl_trusted_certificate /certs/centaur.cloud.crt;
    }
}
```

# Check whether OCSP stapling is enabled

```sh
openssl s_client -connect tc.centaur.cloud:443 -tls1 -tlsextdebug -status < /dev/null 2>&1 | awk '{ if ($0 ~ /OCSP response: no response sent/) { print "disabled" } else if ($0 ~ /OCSP Response Status: successful/) { print "enabled" } }'
```

# Test the OCSP of your certificate

Prepare certs and the OCSP host:

```sh
# Get server cert
openssl s_client -connect tc.centaur.cloud:443 < /dev/null 2>&1 | sed -n '/-----BEGIN/,/-----END/p' > certificate.pem
# Get intermediate cert
openssl s_client -showcerts -connect tc.centaur.cloud:443 < /dev/null 2>&1 | sed -n '/-----BEGIN/,/-----END/p' | awk 'BEGIN { n=0 } { if ($0=="-----BEGIN CERTIFICATE-----") { n+=1 } if (n>=2) { print $0 } }' > chain.pem
# Get the OCSP responder for server cert
openssl x509 -noout -ocsp_uri -in certificate.pem
# http://ocsp.int-x3.letsencrypt.org
```

Make the OCSP request:

```sh
openssl ocsp -issuer chain.pem -cert certificate.pem \
        -verify_other chain.pem \
        -header "Host" "ocsp.int-x3.letsencrypt.org" -text \
        -url http://ocsp.int-x3.letsencrypt.org
```

result:

```
OCSP Request Data:
    Version: 1 (0x0)
    Requestor List:
        Certificate ID:
          Hash Algorithm: sha1
          Issuer Name Hash: 7EE66AE7729AB3FCF8A220646C16A12D6071085D
          Issuer Key Hash: A84A6A63047DDDBAE6D139B7A64565EFF3A8ECA1
          Serial Number: 0353F3B3D1D03160B982105841C733978C28
    Request Extensions:
        OCSP Nonce: 
            0410104F5B81F58C45149ACD6EF72B64A333
OCSP Response Data:
    OCSP Response Status: successful (0x0)
    Response Type: Basic OCSP Response
    Version: 1 (0x0)
    Responder Id: C = US, O = Let's Encrypt, CN = Let's Encrypt Authority X3
    Produced At: Jul 15 00:19:00 2020 GMT
    Responses:
    Certificate ID:
      Hash Algorithm: sha1
      Issuer Name Hash: 7EE66AE7729AB3FCF8A220646C16A12D6071085D
      Issuer Key Hash: A84A6A63047DDDBAE6D139B7A64565EFF3A8ECA1
      Serial Number: 0353F3B3D1D03160B982105841C733978C28
    Cert Status: good
    This Update: Jul 15 00:00:00 2020 GMT
    Next Update: Jul 22 00:00:00 2020 GMT

    Signature Algorithm: sha256WithRSAEncryption
         1d:a8:35:ba:14:83:fe:1a:0b:95:e8:b8:9f:5a:18:3f:fa:ca:
         3b:db:74:10:68:9c:dd:aa:e2:c3:af:2e:c7:c6:80:02:49:84:
         e1:4b:98:0f:b4:e1:88:4d:14:7d:ae:18:12:ee:0c:21:6d:c0:
         7a:00:48:17:a2:b9:0b:80:34:34:cb:00:a0:cf:ee:86:c0:ea:
         6d:66:0e:eb:af:0c:30:93:4f:c1:86:46:15:e1:5f:60:3d:5f:
         33:dc:3e:97:a5:8d:94:52:b9:b1:fe:1a:0a:b1:59:4b:a2:d2:
         11:fe:09:87:9e:ce:5f:c7:8b:b5:3c:c0:a2:61:a8:37:0b:93:
         3c:0b:82:2e:da:49:76:4a:23:e2:4d:45:4b:81:34:90:8d:0c:
         a0:65:76:8a:de:0f:32:bb:1f:da:fa:91:32:d2:c3:4a:d5:d8:
         04:66:ec:1d:d3:12:12:a6:6a:23:93:6e:d1:45:c7:12:ce:7a:
         0a:c8:47:31:fc:1f:e3:19:a2:c0:02:2a:26:55:a6:58:7b:41:
         31:1c:6e:55:cf:68:08:b3:05:dd:96:31:15:bb:14:9b:7c:65:
         e6:18:de:fa:1a:9d:59:7a:b1:41:fc:d7:88:8c:5e:56:9f:c7:
         69:f8:2f:be:6c:ae:0c:7f:9a:58:d1:39:c3:55:1a:5f:2c:42:
         c8:3b:20:14
WARNING: no nonce in response
Response verify OK
certificate.pem: good
    This Update: Jul 15 00:00:00 2020 GMT
    Next Update: Jul 22 00:00:00 2020 GMT
```

if the OCSP host has been blocked by a firewall, you will hang after `OCSP Request Data` and wait `OCSP Response Data` for a long time.

> The content above reference from https://akshayranganath.github.io/OCSP-Validation-With-Openssl/

