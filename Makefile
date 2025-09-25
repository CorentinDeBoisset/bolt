ifeq ($(VERSION),)
	VERSION =
endif

PACKAGE_NAME := bolt
BIN_DIR := bin
GO_FILES := $(shell find . -type f -name '*.go')

RELEASE_BUILD_FLAGS = -ldflags "-X main.Version=${VERSION} -s -w"

.PHONY: all
all: ${BIN_DIR}/${PACKAGE_NAME}

.PHONY: dev
dev: ${BIN_DIR}/${PACKAGE_NAME}_dev


${BIN_DIR}/${PACKAGE_NAME}_dev: ${GO_FILES}
	go build -o $@

# Standard build, for the current platform
${BIN_DIR}/${PACKAGE_NAME}: ${GO_FILES}
	go build ${RELEASE_BUILD_FLAGS} -o $@

# Cross-platform builds
.PHONY: cross-platform
cross-platform: \
	${BIN_DIR}/${PACKAGE_NAME}_linux_amd64 \
	${BIN_DIR}/${PACKAGE_NAME}_linux_arm64 \
	${BIN_DIR}/${PACKAGE_NAME}_darwin_amd64 \
	${BIN_DIR}/${PACKAGE_NAME}_darwin_arm64

${BIN_DIR}/${PACKAGE_NAME}_linux_amd64: ${GO_FILES}
	GOOS=linux GOARCH=amd64 go build ${RELEASE_BUILD_FLAGS} -o $@
${BIN_DIR}/${PACKAGE_NAME}_linux_arm64: ${GO_FILES}
	GOOS=linux GOARCH=arm64 go build ${RELEASE_BUILD_FLAGS} -o $@

${BIN_DIR}/${PACKAGE_NAME}_darwin_amd64: ${GO_FILES}
	GOOS=darwin GOARCH=amd64 go build ${RELEASE_BUILD_FLAGS} -o $@
${BIN_DIR}/${PACKAGE_NAME}_darwin_arm64: ${GO_FILES}
	GOOS=darwin GOARCH=arm64 go build ${RELEASE_BUILD_FLAGS} -o $@


.PHONY: clean
clean:
	rm -f ${BIN_DIR}/${PACKAGE_NAME}*

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
