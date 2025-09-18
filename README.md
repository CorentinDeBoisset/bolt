# Localci

Local-CI is an orchestrator to execute complex CI jobs locally.

## ❯ Installation steps

You can use a pre-built binary with:

```bash
wget https://github.com/CorentinDeBoisset/localci/releases/download/<version>/localci_<platform>_<arch> -O localci

# On linux, you can install on /usr/local/bin (requires sudo), or in ~/bin
install -m 0755 localci /path/to/install
```

Alternatively, you can run the following command (you will need to install golang v1.25+):

```bash
go install github.com/corentindeboisset/localci
```

## ❯ Contributing

### Development setup

You can build the development binary:

```bash
make dev
./bin/localci_dev <args>
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
