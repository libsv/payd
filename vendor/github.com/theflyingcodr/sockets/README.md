# sockets

> Channel based socket library on top of gorilla/sockets giving listeners, middleware, server and clients.

[![Release](https://img.shields.io/github/release-pre/theflyingcodr/sockets.svg?logo=github&style=flat&v=1)](https://github.com/theflyingcodr/sockets/releases)
[![Build Status](https://github.com/theflyingcodr/sockets/actions/workflows/go.yml/badge.svg)](https://github.com/theflyingcodr/sockets/actions/workflows/go.yml)
[![CodeQL Status](https://github.com/theflyingcodr/sockets/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/theflyingcodr/sockets/actions/workflows/codeql-analysis.yml)
[![Report](https://goreportcard.com/badge/github.com/theflyingcodr/sockets?style=flat&v=1)](https://goreportcard.com/report/github.com/theflyingcodr/sockets)
[![codecov](https://codecov.io/gh/libsv/go-bt/branch/master/graph/badge.svg?v=1)](https://codecov.io/gh/theflyingcodr/sockets)
[![Go](https://img.shields.io/github/go-mod/go-version/theflyingcodr/sockets?v=1)](https://golang.org/)

## Table of Contents

- [Installation](#installation)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

## Installation

**sockets** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).

```shell script
go get -u github.com/theflyingcodr/sockets
```

## Documentation

View the generated [documentation](https://pkg.go.dev/github.com/theflyingcodr/sockets)

[![GoDoc](https://godoc.org/github.com/theflyingcodr/sockets?status.svg&style=flat)](https://pkg.go.dev/github.com/theflyingcodr/sockets)

For more information around the technical aspects of Bitcoin, please see the updated [Bitcoin Wiki](https://wiki.bitcoinsv.io/index.php/Main_Page)

### Features

- Client / Server implementations
- Middleware
- Error handlers
- Listener funcs
- Channel based messaging
- Client reconnects
- Broadcasting directly from server to client or to entire channel

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to Github and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.
</details>

## Examples & Tests

All unit tests and [examples](examples) run via [Github Actions](https://github.com/theflyingcodr/sockets/actions) and
uses [Go version 1.17.x](https://golang.org/doc/go1.17). View the [configuration file](.github/workflows/go.yml).

Run all tests

```shell script
make pre-commit
```

## Usage

View the [examples](examples)

<br/>

## Maintainers

| [<img src="https://github.com/theflyingcodr.png" height="50" alt="JH" />](https://github.com/theflyingcodr) 
|:---:|
|  [MS](https://github.com/theflyingcodr)| 

<br/>

## Contributing

View the [contributing guidelines](CONTRIBUTING.md) and please follow the [code of conduct](CODE_OF_CONDUCT.md).

### How can I help?

All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/theflyingcodr) :clap:

<br/>

## License

![License](https://img.shields.io/github/license/theflyingcodr/sockets.svg?style=flat&v=1)
