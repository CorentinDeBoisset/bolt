# Localci

Local-CI is an orchestrator to execute complex CI jobs locally.

## ❯ Installation steps

Requirements:

- golang v1.23+

Run the following command:

    go install github.com/corentindeboisset/localci

In the future, some precompiled binaries may be automatically built.

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
