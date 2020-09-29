BINARY_DEST_DIR ?= bin

GO_BUILD_OPTIONS ?= -buildmode=pie

.PHONY: build-cli
build-cli:
	GOFLAGS=-mod=vendor go build ${GO_BUILD_OPTIONS} -o ${BINARY_DEST_DIR}/monitor github.com/yangzuo0621/monitor/cmd/monitor
