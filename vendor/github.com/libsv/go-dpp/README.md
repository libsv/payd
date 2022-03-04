# DPP - Direct Payment Protocol

[![Release](https://img.shields.io/github/release-pre/libsv/go-p4.svg?logo=github&style=flat&v=1)](https://github.com/libsv/go-dpp/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/go-p4/run-go-tests?logo=github&v=3)](https://github.com/libsv/go-dpp/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/go-dpp?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/go-dpp)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/go-p4?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)

DPP is a basic reference implementation of a Payment Protocol Server implementing the proposed BIP-270 payment flow.

This is written in go and integrates with a wallet running the Payment Protocol PayD Interface.

## Exploring Endpoints

To explore the endpoints and functionality, run the server using `go run cmd/rest-server/main.go` and navigate to [Swagger](http://localhost:8445/swagger/index.html) 
where the endpoints and their models are described in detail.

## Configuring DPP

The server has a series of environment variables that allow you to configure the behaviours and integrations of the server.
Values can also be passed at build time to provide information such as build information, region, version etc.

### Server

| Key                    | Description                                                        | Default        |
| ---------------------- | ------------------------------------------------------------------ | -------------- |
| SERVER_PORT            | Port which this server should use                                  | :8445          |
| SERVER_HOST            | Host name under which this server is found                         | dpp            |
| SERVER_SWAGGER_ENABLED | If set to true we will expose an endpoint hosting the Swagger docs | true           |
| SERVER_SWAGGER_HOST    | Sets the base url for swagger ui calls                             | localhost:8445 |

### Environment / Deployment Info

| Key                 | Description                                                                | Default          |
| ------------------- | -------------------------------------------------------------------------- | ---------------- |
| ENV_ENVIRONMENT     | What enviornment we are running in, for example 'production'               | dev              |
| ENV_REGION          | Region we are running in, for example 'eu-west-1'                          | local            |
| ENV_COMMIT          | Commit hash for the current build                                          | test             |
| ENV_VERSION         | Semver tag for the current build, for example v1.0.0                       | v0.0.0           |
| ENV_BUILDDATE       | Date the code was build                                                    | Current UTC time |
| ENV_BITCOIN_NETWORK | What bitcoin network we are connecting to (mainnet, testnet, stn, regtest) | regtest          |

### Logging

| Key       | Description                                                           | Default |
| --------- | --------------------------------------------------------------------- | ------- |
| LOG_LEVEL | Level of logging we want within the server (debug, error, warn, info) | info    |

### PayD Wallet

| Key         | Description                                               | Default |
| ----------- | --------------------------------------------------------- | ------- |
| PAYD_HOST   | Host for the wallet we are connecting to                  | payd    |
| PAYD_PORT   | Port the PayD wallet is listening on                      | :8443   |
| PAYD_SECURE | If true the DPP server will validate the wallet TLS certs | false   |
| PAYD_NOOP   | If true we will use a dummy data store in place of payd   | true    |

## Working with DPP

There are a set of makefile commands listed under the [Makefile](Makefile) which give some useful shortcuts when working
with the repo.

Some of the more common commands are listed below:

`make pre-commit` - ensures dependencies are up to date and runs linter and unit tests.

`make build-image` - builds a local docker image, useful when testing dpp in docker.

`make run-compose` - runs DPP in compose, a reference PayD wallet will be added to compose soon NOTE the above command will need ran first.

### Rebuild on code change

You can also add an optional `docker-compose.dev.yml` file (this is not committed) where you can safely overwrite values or add other services without impacting the main compose file.

If you add this file, you can run it with `make run-compose-dev`.

The file I use has a watcher which means it auto rebuilds the image on code change and ensures compose is always up to date, this full file is shown below:

```yaml
version: "3.7"

services:
  dpp:
    image: theflyingcodr/go-watcher:1.15.8
    environment:
      GO111MODULE: "on"
      GOFLAGS: "-mod=vendor"
      DB_DSN: "file:data/wallet.db?cache=shared&_foreign_keys=true;"
      DB_SCHEMA_PATH: "data/sqlite/migrations"
    command: watcher -run github.com/libsv/go-dpp/cmd/rest-server/ -watch github.com/libsv/go-dpp
    working_dir: /go/src/github.com/libsv/go-dpp
    volumes:
      - ~/git/libsv/go-dpp:/go/src/github.com/libsv/go-dpp
```

## CI / CD

We use github actions to test and build the code.

If a new release is required, after your PR is approved and code added to master, simply add a new semver tag and a GitHub action will build and publish your code as well as create a GitHub release.
