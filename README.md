# 🌲 Tera

Boost your development workflow by at least `10^12`.

Tera is a simple CLI tool that help you manage interdependent services, and can run series of tasks without hassle.
This tool was created specifically to simplify the management of local development servers, or the launching of test routines, but you can use it for many other use cases.

It was made possible by the TUI libraries built by the people at [Charmbracelet](https://charm.land/).

## ❯ Installation steps

You can use a pre-built binary with:

```bash
wget "https://github.com/CorentinDeBoisset/tera/releases/latest/download/tera_$(uname)_$(uname -m)" -O tera

# On linux, you can install on /usr/local/bin (requires sudo), or in ~/bin
install -m 0755 tera /path/to/install
```

Alternatively, you can run the following command (you will need to install golang v1.26+):

```bash
go install github.com/corentindeboisset/tera
```

## ❯ Contributing

### Development setup

You can build the development binary:

```bash
make dev
./bin/tera_dev <args>
```

### Tests

If you want to run the tests, you can execute:

```
make test
```

Aditionnaly, if you want a coverage report:

```
make coverage
```

### Translation management

Install the `gotext` excutable:

```bash
go install golang.org/x/text/cmd/gotext@latest
```

Then update the translation catalogs with:

```bash
go generate ./...
```
