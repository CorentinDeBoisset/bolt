VERSION = 1.0.0

PACKAGE_NAME := localci
BIN_DIR := bin
GO_FILES := $(shell find . -type f -name '*.go')

DEV_BUILD_FLAGS = -ldflags "-X main.Version=${VERSION}"
RELEASE_BUILD_FLAGS = -ldflags "-X main.Version=${VERSION} -s -w"

.PHONY: all
all: ${BIN_DIR}/${PACKAGE_NAME}

.PHONY: dev
dev: ${BIN_DIR}/${PACKAGE_NAME}_dev

${BIN_DIR}/${PACKAGE_NAME}: ${GO_FILES}
	go build ${RELEASE_BUILD_FLAGS} -o $@

${BIN_DIR}/${PACKAGE_NAME}_dev: ${GO_FILES}
	go build ${DEV_BUILD_FLAGS} -o $@

.PHONY: clean
clean:
	rm -f ${BIN_DIR}/${PACKAGE_NAME} ${BIN_DIR}/${PACKAGE_NAME}_dev

.PHONY: install
install:
	go install ${RELEASE_BUILD_FLAGS}

.PHONY: test
test:
	go test ./...

_cover.out: ${GO_FILES}
	go test ./... -coverprofile=_cover.out

.PHONY: coverage
coverage: _cover.out
	go tool cover -html=_cover.out
