# ⛈️ Bolt

Bolt is a task orchestrator to execute complex jobs.

## ❯ Installation steps

You can use a pre-built binary with:

```bash
wget "https://github.com/CorentinDeBoisset/bolt/releases/download/v1.6.0/bolt_$(uname)_$(uname -m)" -O bolt

# On linux, you can install on /usr/local/bin (requires sudo), or in ~/bin
install -m 0755 bolt /path/to/install
```

Alternatively, you can run the following command (you will need to install golang v1.25+):

```bash
go install github.com/corentindeboisset/bolt
```

## ❯ Contributing

### Development setup

You can build the development binary:

```bash
make dev
./bin/bolt_dev <args>
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
