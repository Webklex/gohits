# GoHits visitor counter

[![Hits][link-hits]][link-website]
[![Total Downloads][link-downloads]][link-releases]
[![Latest Stable Version][link-version]][link-releases]
[![License][link-version]](LICENSE.md)
[![Website status][link-status]][link-website]

## Description
An easy way to track your page or project views ("hits") of any GitHub or online project.

Generate your own: [hits.webklex.com](https://hits.webklex.com)

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Configuration](#server-options)
  - [HTTP & HTTPS](#http--https)
  - [Letsencrypt](#letsencrypt)
  - [Middlewares & Extensions](#middlewares--extensions)
  - [Rate limiting & Quota management](#rate-limiting--quota-management)
  - [Logging](#logging)
  - [Additional](#additional)
- [Api](#api)
  - [Output](#output)
    - [CSV](#csv)
    - [XML](#xml)
    - [JSON](#json)
  - [Websocket](#websocket)
- [Build](#build)
- [Support](#support)
- [Security](#security)
- [Credits](#credits)
- [License](#license)

### Features
* Serving over HTTPS (TLS) using your own certificates, or provisioned automatically using [LetsEncrypt.org](https://letsencrypt.org)
* [HSTS ready](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) to restrict your browser clients to always use HTTPS
* Configurable read and write timeouts to avoid stale clients consuming server resources
* Reverse proxy ready
* Configurable [CORS](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) to restrict access to specific domains
* Configurable api prefix to serve the API alongside other APIs on the same host
* Optional round trip optimization by enabling [TCP Fast Open](https://en.wikipedia.org/wiki/TCP_Fast_Open)
* Integrated rate limit (quota) for your clients (per client IP) based on requests per time interval; several backends such as in-memory map (for single instance), or redis or memcache for distributed deployments are supported

### Installation
Download and unpack a fitting [pre-compiled binary](https://github.com/webklex/gohits/releases) or build a binary 
yourself by by following the [build](#build) instructions.

Continue by configuring your application:
```bash
gohits -http=:8080 -gui gui -save
```
Open a browser and navigate to `http://localhost:8080/` to verify everything is working.

Please take a look at the available [options](#server-options) for further details.

### Server Options
To see all the available options, use the `-help` option:
```bash
gohits -help
```
                        |

#### HTTP & HTTPS
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -http                  | HTTP                 | string | localhost:8080       | Address in form of ip:port to listen                        |
| -https                 | HTTPS                | string |                      | Address in form of ip:port to listen                        |
| -write-timeout         | WRITE_TIMEOUT        | int    | 15000000000          | Write timeout in nanoseconds for HTTP and HTTPS client connections |
| -read-timeout          | READ_TIMEOUT         | int    | 30000000000          | Read timeout in nanoseconds for HTTP and HTTPS client connections |
| -tcp-fast-open         | TCP_FAST_OPEN        | bool   | false                | Enable TCP fast open                                        |
| -tcp-naggle            | TCP_NAGGLE           | bool   | false                | Enable TCP Nagle's algorithm                                |
| -http2                 | HTTP2                | bool   | true                 | Enable HTTP/2 when TLS is enabled                           |
| -hsts                  | HSTS                 | string |                      |                                                             |
| -key                   | KEY                  | string | key.pem              | X.509 key file for HTTPS server                             |
| -cert                  | CERT                 | string | cert.pem             | X.509 certificate file for HTTPS server                     |

#### Letsencrypt
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -letsencrypt           | LETSENCRYPT          | bool   | false                | Enable automatic TLS using letsencrypt.org                  |
| -letsencrypt-email     | LETSENCRYPT_EMAIL    | string |                      | Optional email to register with letsencrypt                 |
| -letsencrypt-hosts     | LETSENCRYPT_HOSTS    | string |                      | Comma separated list of hosts for the certificate           |
| -letsencrypt-cert-dir  | LETSENCRYPT_CERT_DIR | string |                      | Letsencrypt cert dir                                        |

#### Middlewares & Extensions
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -use-x-forwarded-for   | USE_X_FORWARDED_FOR  | bool   | false                | Use the X-Forwarded-For header when available (e.g. behind proxy) |
| -cors-origin           | CORS_ORIGIN          | string | *                    | Comma separated list of CORS origins endpoints              |
| -api-prefix            | API_PREFIX           | string | /                    | API endpoint prefix                                         |
| -gui                   | GUI                  | string |                      | Web gui directory                                           |
| -session-lifetime      | SESSION_LIFETIME     | int    | 1200000000000        | Session lifetime of an counted visitor (default 20min)      |
| -pong-wait             | PONG_WAIT            | int    | 24000000000          | Time allowed to read the next pong message from the peer. (default 24s) |
| -ping-period           | PING_PERIOD          | int    | 12000000000          | Send pings to peer with this period. Must be less than pong-wait. (default 12s) |

##### Rate limiting & Quota management
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -quota-burst           | QUOTA_BURST          | int    | 3                    | Max requests per source IP per request burst                |
| -quota-interval        | QUOTA_INTERVAL       | int    | 3600000000000        | Quota expiration interval, per source IP querying the API in nanoseconds |
| -quota-max             | QUOTA_MAX            | int    | 1                    | "Max requests per source IP per interval; set 0 to turn quotas off |

#### Logging
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -logtostdout           | LOGTOSTDOUT          | bool   | false                | Log to stdout instead of stderr                             |
| -log-file              | LOG_FILE             | string |                      | Log file location                             |
| -logtimestamp          | LOGTIMESTAMP         | bool   | true                 | Prefix non-access logs with timestamp                       |

#### Additional
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -silent                | SILENT               | bool   | false                | Disable HTTP and HTTPS log request details                  |
| -config                |                      | string | conf/settings.config | Config file path                                            |
| -save                  |                      | bool   | false                | Save config                                                 |
| -version               |                      | bool   | false                | Show version and exit                                       |
| -help                  |                      | bool   | false                | Show help and exit                                          |

If you're using LetsEncrypt.org to provision your TLS certificates, you have to listen for HTTPS on port 443. Following 
is an example of the server listening on 2 different ports: http (80) and https (443):
```bash
gohits \
    -http=:8080 \
    -https=:8443 \
    -hsts=max-age=31536000 \
    -letsencrypt \
    -letsencrypt-hosts=example.com \
    -gui gui \
    -save
```

```bash
$ cat conf/settings.config
{
    "HTTP": ":8080",
    "HTTPS": ":8443",
    "HSTS": "max-age=31536000",
    "LETSENCRYPT": true,
    "LETSENCRYPT_HOSTS": "example.com",
    ...
```

By default, HTTP/2 is enabled over HTTPS. You can disable by passing the `-http2=false` flag.

If the web server is running behind a reverse proxy or load balancer, you have to run it passing the `-use-x-forwarded-for` 
parameter and provide the `X-Forwarded-For` HTTP header in all requests. This is for the gohits web server be able to log the 
client IP, and to perform correctly identify new hits.

## API
The API is served by endpoints that encode the response in different formats.

```bash
curl :8080/json/{username}/{repository}
```
Same semantics are available for the `/xml/{username}/{repository}` and `/csv/{username}/{repository}` endpoints.

### Output
#### Section
| Name                  | Value type    | JSON                      | XML                   | CSV   |
| :-------------------- | :------------ | :------------------------ | :-------------------- | :---- |
| Username              | string        | username                  | Username              | 0     |
| Repository            | string        | repository                | Repository            | 1     |
| Total                 | int           | total                     | Total                 | 2     |
| Created at            | datetime      | created_at                | CreatedAt             | 3     |
| Updated at            | datetime      | updated_at                | UpdatedAt             | 4     |

#### CSV
```bash
curl :8080/csv/webklex/gohits
```
```
webklex,gohits,55,2020-09-11 07:01:23,2020-09-12 00:10:07
```

#### XML
```bash
curl :8080/xml/webklex/gohits
```
```xml
<Section>
    <Username>webklex</Username>
    <Repository>gohits</Repository>
    <Total>55</Total>
    <CreatedAt>2020-09-11T07:01:23.252745204+02:00</CreatedAt>
    <UpdatedAt>2020-09-12T00:10:07.7275806+02:00</UpdatedAt>
</Section>
```

#### JSON
```bash
curl :8080/json/webklex/gohits
```
```json
{
  "username": "webklex",
  "repository": "gohits",
  "total": 55,
  "created_at": "2020-09-11T07:01:23.252745204+02:00",
  "updated_at": "2020-09-12T00:10:07.7275806+02:00"
}
```

### Websocket
Url: `:8080/ws`

You can subscribe to specific channels or to `all` in order to receive all recent hits.

#### Payloads:
**Subscribe to channel `all`:**
```json
{
  "name": "subscribe",
  "payload": "all"
}
```
**Subscribe to channel `webklex/gohits`:**
```json
{
  "name": "subscribe",
  "payload": "webklex/gohits"
}
```
**Delete a the subscription of `all`:**
```json
{
  "name": "unsubscribe",
  "payload": "all"
}
```
#### Output
```
00:26:07 webklex/gohits
```


### Build
You can build your own binaries by calling `build.sh`
```bash
build.sh build_dir
```

### Features & pull requests
Everyone can contribute to this project. Every pull request will be considered but it can also happen to be declined. 
To prevent unnecessary work, please consider to create a [feature issue](https://github.com/webklex/gohits/issues/new?template=feature_request.md) 
first, if you're planning to do bigger changes. Of course you can also create a new [feature issue](https://github.com/webklex/gohits/issues/new?template=feature_request.md)
if you're just wishing a feature ;)

>Off topic, rude or abusive issues will be deleted without any notice.


## Support
If you encounter any problems or if you find a bug, please don't hesitate to create a new [issue](https://github.com/webklex/gohits/issues).
However please be aware that it might take some time to get an answer.

If you need **immediate** or **commercial** support, feel free to send me a mail at github@webklex.com. 

## Change log

Please see [CHANGELOG](CHANGELOG.md) for more information what has changed recently.

## Security

If you discover any security related issues, please email github@webklex.com instead of using the issue tracker.

## Credits
- [Webklex][link-author]
- [All Contributors][link-contributors]

## License
The MIT License (MIT). Please see [License File](LICENSE.md) for more information.

[link-downloads]: https://img.shields.io/github/downloads/webklex/gohits/total?style=flat-square
[link-version]: https://img.shields.io/github/license/webklex/gohits?style=flat-square
[link-license]: https://hits.webklex.com
[link-website]: https://hits.webklex.com
[link-releases]: https://github.com/webklex/gohits/releases
[link-hits]: https://hits.webklex.com/svg/webklex/gohits
[link-author]: https://github.com/webklex
[link-contributors]: https://github.com/webklex/gohits/graphs/contributors
[link-status]: https://img.shields.io/website?down_message=Offline&label=Website&style=flat-square&up_message=Online&url=https%3A%2F%2Fhits.webklex.com%2F