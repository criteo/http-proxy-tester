# http-proxy-tester

http-proxy-tester makes request to HTTP(S) targets via a proxy using [HTTP Basic authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) and will exit(1) if all requests failed; exit(0) if at least 1 request was successful.

## Getting started

You should have a working Golang environment [setup](https://golang.org/doc/install).

```
go get github.com/criteo/http-proxy-tester
```

### Build

```
cd $GOPATH/src/github.com/criteo/http-proxy-tester/
make build
```

### Install

```
cd $GOPATH/src/github.com/criteo/http-proxy-tester/
make
```

### Running http-proxy-tester

The application requires a configuration file (see the [example configuration file](config.example.yml)). By default, it will check for a `config` file in the current directory.

```
./http-proxy-tester
```

# License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.