#############
# VARIABLES #
#############

BUILD_DATE := $(shell date -u '+%Y-%m-%d')
GIT_COMMIT := $(shell git rev-parse --short HEAD || echo "unknown")
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/-dirty//' | grep v || echo "v0.0.0-$(GIT_COMMIT)")
LOCALBIN ?= $(shell pwd)/bin
# Version information for the build
LDFLAGS := -X github.com/kagent-dev/tools/internal/version.Version=$(VERSION) -X github.com/kagent-dev/tools/internal/version.GitCommit=$(GIT_COMMIT) -X github.com/kagent-dev/tools/internal/version.BuildDate=$(BUILD_DATE)


#########
# UTILS #
#########

clean:
	rm -rf ./bin/kyverno-agent-tools-*
	rm -rf $(HOME)/.local/bin/kyverno-agent-tools-*

###############
# LINUX AMD64 #
###############
bin/kyverno-agent-tools-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/kyverno-agent-tools-linux-amd64 ./cmd

bin/kyverno-agent-tools-linux-amd64.sha256: bin/kyverno-agent-tools-linux-amd64
	sha256sum bin/kyverno-agent-tools-linux-amd64 > bin/kyverno-agent-tools-linux-amd64.sha256

################
# LINUX ARM64 #
################
bin/kyverno-agent-tools-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/kyverno-agent-tools-arm64 ./cmd

bin/kyverno-agent-tools-arm64.sha256: bin/kyverno-agent-tools-arm64
	sha256sum bin/kyverno-agent-tools-arm64 > bin/kyverno-agent-tools-arm64.sha256

################
# DARWIN ARM64 #
################
bin/kyverno-agent-tools-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/kyverno-agent-tools-darwin-arm64 ./cmd

bin/kyverno-agent-tools-darwin-arm64.sha256: bin/kyverno-agent-tools-darwin-arm64
	sha256sum bin/kyverno-agent-tools-darwin-arm64 > bin/kyverno-agent-tools-darwin-arm64.sha256


.PHONY: build
build: $(LOCALBIN) clean bin/kyverno-agent-tools-linux-amd64.sha256 bin/kyverno-agent-tools-arm64.sha256 bin/kyverno-agent-tools-darwin-arm64.sha256
build:
	@echo "Build complete. Binaries are available in the bin/ directory."
	ls -lt bin/kyverno-agent-tools-*

.PHONY: install
install: clean
	mkdir -p $(HOME)/.local/bin
	go build -ldflags "$(LDFLAGS)" -o $(LOCALBIN)/kyverno-agent-tools ./cmd
	go build -ldflags "$(LDFLAGS)" -o $(HOME)/.local/bin/kyverno-agent-tools ./cmd

.PHONY: $(LOCALBIN)
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
