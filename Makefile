BINARY_DEST_DIR ?= bin

.PHONY: build-cli
build-cli:
	go build ${GO_BUILD_OPTIONS} -o ${BINARY_DEST_DIR}/monitor github.com/yangzuo0621/azure-devops-cmd/cmd/monitor
