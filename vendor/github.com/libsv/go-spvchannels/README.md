# go-spvchannels

[![Release](https://img.shields.io/github/release-pre/libsv/go-spvchannels.svg?logo=github&style=flat&v=1)](https://github.com/libsv/go-spvchannels/releases)
[![Go](https://github.com/libsv/go-bt/actions/workflows/run-tests.yml/badge.svg?branch=master)](https://github.com/libsv/go-bt/actions/workflows/run-tests.yml)
[![Go](https://github.com/libsv/go-bt/actions/workflows/run-tests.yml/badge.svg?event=pull_request)](https://github.com/libsv/go-bt/actions/workflows/run-tests.yml)

[![Report](https://goreportcard.com/badge/github.com/libsv/go-spvchannels?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/go-spvchannels)
[![codecov](https://codecov.io/gh/libsv/go-spvchannels/branch/master/graph/badge.svg?v=1)](https://codecov.io/gh/libsv/go-spvchannels)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/go-spvchannels?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)
[![Mergify Status][mergify-status]][mergify]

[mergify]: https://mergify.io
[mergify-status]: https://img.shields.io/endpoint.svg?url=https://gh.mergify.io/badges/libsv/go-spvchannels&style=flat
<br/>

Go SPV Channels is a [golang](https://golang.org/) implementation of the [SPV Channels Server](https://github.com/bitcoin-sv/spvchannels-reference).

The library implement all rest api endpoints served in [SPV Channels Server](https://github.com/bitcoin-sv/spvchannels-reference) and the websocket client to listen new message notifications in real time.

## Table of Contents

- [Installation](#installation)
- [Run tests](#run-tests)
- [Setup Local SPV Channels server](#setup-local-spv-channels-server)
- [License](#license)

## Installation
```
go get github.com/libsv/go-spvchannels
```

## Run tests

Run unit test
```
go clean -testcache && go test -v ./...
```

To run integration tests, make sure you have `docker-compose up -d` on your local machine, then run
```
go clean -testcache && go test  -race -v -tags=integration ./...
```

## Setup Local SPV Channels server

#### Creating SSL key for secure connection
Following the tutorial in the [SPV Channels Server](https://github.com/bitcoin-sv/spvchannels-reference), we first create the certificate using `openssl`:
```
terminal $> openssl req -x509 -out localhost.crt -keyout localhost.key -newkey rsa:2048 -nodes -sha256 -subj '/CN=localhost' -extensions EXT -config <( printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
terminal $> openssl pkcs12 -export -out devkey.pfx -inkey localhost.key -in localhost.crt # use devkey as password
```

That will create the `devkey.pfx` with password `devkey`. We then write a `docker-compose.yml` file following the tutorial.

#### Launch local [SPV Channels Server](https://github.com/bitcoin-sv/spvchannels-reference) :
```
docker-compose up -d
```

#### Create an account
We then need to create a SPV Channels account on the server
```
docker exec spvchannels ./SPVChannels.API.Rest -createaccount spvchannels_dev dev dev
```

#### Usage with swagger

The [SPV Channels Server](https://github.com/bitcoin-sv/spvchannels-reference) run by `docker-compose.yml` listen on `localhost:5010`. We can start playing with the endpoints using swagger, i.e in browser, open `https://localhost:5010/swagger/index.html`

From this page, there are a link `/swagger/v1/swagger.json` to export swagger file

#### Usage with Postman

Interacting with browser might have some difficulty related to adding certificate to the system. It might be easier to use Postman to interact as Postman has a easy possibility to disable SSL certificate check to ease development propose.

From Postman, import the file `devconfig/postman.json` and set the environment config as follow

| VARIABLE    | INITIAL VALUE  |
| ----------- | -------------- |
| URL_PORT    | localhost:5010 |
| ACCOUNT     | 1              |
| USERNAME    | dev            |
| PASSWORD    | dev            |

These environment variable are used as _template_ to populate values in the `postman.json` file. There are a few more environment variable to define (look into the json file) that will depend to the endpoint and value created during the experience:

| VARIABLE     | INITIAL VALUE   |
| ------------ | --------------- |
| CHANNEL_ID   | .. to define .. |
| TOKEN_ID     | .. to define .. |
| TOKEN_VALUE  | .. to define .. |
| MSG_SEQUENCE | .. to define .. |
| NOTIFY_TOKEN | .. to define .. |

## License

![License](https://img.shields.io/github/license/libsv/go-spvchannels.svg?style=flat&v=1)