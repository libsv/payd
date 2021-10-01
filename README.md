# payd 

[![Release](https://img.shields.io/github/release-pre/libsv/payd.svg?logo=github&style=flat&v=1)](https://github.com/libsv/payd/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/payd/run-go-tests?logo=github&v=3)](https://github.com/libsv/payd/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/payd?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/payd)
[![codecov](https://codecov.io/gh/libsv/go-bt/branch/master/graph/badge.svg?v=1)](https://codecov.io/gh/libsv/payd)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/payd?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)

Payd is a basic dummy wallet (do not use this) for demonstrating the BIP 270 / Payment Protocol flow.

It has a random master key, created at startup and a single user support for now and no authentication. Seriously, don't use this wallet at the moment expect for demonstration purposes.

This wallet has an Invoice interface with CRUD operations for creating payment invoices and also implements the Wallet Payment Protocol Interface, used to integration with payment protocol servers.

This is written in go and integrates with servers running the Payment Protocol Interface.

## Exploring Endpoints

To explore the endpoints and functionality, run the server using `go run cmd/rest-server/main.go` and navigate to [Swagger](http://localhost:8443/swagger/index.html)
where the endpoints and their models are described in detail.

## Configuring PayD

The server has a series of environment variables that allow you to configure the behaviours and integrations of the server.
Values can also be passed at build time to provide information such as build information, region, version etc.

### Server

| Key                    | Description                                                        | Default       |
|------------------------|--------------------------------------------------------------------|---------------|
| SERVER_PORT            | Port which this server should use                                  | :8443         |
| SERVER_HOST            | Host name under which this server is found                         | payd:8443     |
| SERVER_SWAGGER_ENABLED | If set to true we will expose an endpoint hosting the Swagger docs | true          |
| SERVER_SWAGGER_HOST    | The host that swagger will point its api requests to               | localhost:8443|

### Environment / Deployment Info

| Key                 | Description                                                                | Default          |
|---------------------|----------------------------------------------------------------------------|------------------|
| ENV_ENVIRONMENT     | What enviornment we are running in, for example 'production'               | dev              |
| ENV_REGION          | Region we are running in, for example 'eu-west-1'                          | local            |
| ENV_COMMIT          | Commit hash for the current build                                          | test             |
| ENV_VERSION         | Semver tag for the current build, for example v1.0.0                       | v0.0.0           |
| ENV_BUILDDATE       | Date the code was build                                                    | Current UTC time |

### Logging

| Key       | Description                                                           | Default |
|-----------|-----------------------------------------------------------------------|---------|
| LOG_LEVEL | Level of logging we want within the server (debug, error, warn, info) | info    |

### DB

| Key         | Description                                              | Default |
|-------------|----------------------------------------------------------|---------|
| DB_TYPE   | Type of db you're connecting to (sqlite, postgres,mysql) sqlite only supported currently | sqlite    |
| DB_DSN   | Connection string for the db                     | file:data/wallet.db?_foreign_keys=true&pooled=true   |
| DB_SCHEMA_PATH | Location of the data base migration scripts | data/sqlite/migrations   |
| DB_MIGRATE   | If true we will check the db version and apply missing migrations  | true    |

### Headers Client

If validating using SPV you will need to run a Headers Client, this will sync headers as they are mined and provide 
block and merkle proof information.

| Key         | Description                                              | Default |
|-------------|----------------------------------------------------------|---------|
| HEADERSCLIENT_ADDRESS   | Uri for the headers client you are using | http://headersv:8080    |
| HEADERSCLIENT_TIMEOUT   | Timeout in seconds for headers client queries                     | 30   |

### Wallet

| Key         | Description                                              | Default |
|-------------|----------------------------------------------------------|---------|
| WALLET_NETWORK   | Bitcoin network we're connected to (regtest, stn, testnet,regtest) | regtest    |
| WALLET_SPVREQUIRED   | If true we will require full SPV envelopes to be sent as part of payments | true   |
| WALLET_PAYMENTEXPIRY | Duration in hours that invoices will be valid for | 24   |

## Working with PayD

There are a set of makefile commands listed under the [Makefile](Makefile) which give some useful shortcuts when working
with the repo.

Some of the more common commands are listed below:

`make pre-commit` - ensures dependencies are up to date and runs linter and unit tests.

`make build-image` - builds a local docker image, useful when testing PayD in docker.

`make run-compose` - runs PayD in compose.

## Releases

You can view the latest releases on our [Github Releases](https://github.com/libsv/payd/releases) page.

We also publish docker images which can be found on [Docker Hub](https://hub.docker.com/repository/docker/libsv/payd).

## CI / CD

We use github actions to test and build the code.

If a new release is required, after your PR is approved and code added to master, simply add a new semver tag and a GitHub action will build and publish your code as well as create a GitHub release.

