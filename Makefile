# Makefile for your Go project.

# ---------------------------------------------------------------------------- #
#                             Project Configuration                            #
# ---------------------------------------------------------------------------- #

# List of applications to build/release, corresponding to directories in ./cmd/
# For a single application, just list its name. Example: APPS := my-app
# For multiple, separate with spaces. Example: APPS := my-app-1 my-app-2
# TODO: Replace 'changeme' with the name of your application(s).
APPS := changeme


# ---------------------------------------------------------------------------- #
#                         Tooling & Initialization                         #
# ---------------------------------------------------------------------------- #

.PHONY: init
init: ## 🔧 Install all required development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	# Add other tools here, e.g.:
	# @go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	# @go install github.com/google/wire/cmd/wire@latest


# ---------------------------------------------------------------------------- #
#                       Development Lifecycle Targets                      #
# ---------------------------------------------------------------------------- #

.PHONY: all deps generate clean

all: deps generate test lint ## ✅ Run all essential development steps

deps: ## 📦 Tidy and download Go module dependencies
	@echo "Tidying and downloading dependencies..."
	@go mod tidy
	@go mod download

generate: ## 🧬 Run Go code generation
	@echo "Generating code..."
	@go generate ./...

clean: ## 🧹 Clean up build artifacts and temporary files
	@echo "Cleaning up..."
	@rm -rf ./dist ./bin ./coverage.out


# ---------------------------------------------------------------------------- #
#                             Testing & Linting                              #
# ---------------------------------------------------------------------------- #

.PHONY: test lint

test: ## 🧪 Run all Go tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

lint: ## 🧹 Lint the codebase with golangci-lint
	@echo "Running linter..."
	@golangci-lint run ./...


# ---------------------------------------------------------------------------- #
#                           Build & Release Variables                          #
# ---------------------------------------------------------------------------- #

GOHOSTOS ?= $(shell go env GOHOSTOS)

# Common Git information
GIT_COMMIT      := $(shell git rev-parse --short HEAD)
GIT_BRANCH      := $(shell git rev-parse --abbrev-ref HEAD)
GIT_VERSION     := $(shell git describe --tags --always)
# Get the tag at the current commit. It might be empty.
GIT_HEAD_TAG    := $(shell git tag --points-at HEAD 2>/dev/null)

# OS-specific variables for build date, tree state, and the final version tag
ifeq ($(GOHOSTOS), windows)
    BUILD_DATE   := $(shell powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssK'")
    GIT_TREE_STATE := $(shell powershell -Command "if ((git status) -match 'clean') { 'clean' } else { 'dirty' }")
    # Use the tag if it exists, otherwise use the short commit hash.
    GIT_TAG      := $(shell powershell -Command "if ('${GIT_HEAD_TAG}') { '${GIT_HEAD_TAG}' } else { '${GIT_COMMIT}' }")
else
    BUILD_DATE   := $(shell TZ=Asia/Shanghai date +%FT%T%z)
    # Check for uncommitted changes. git status --porcelain is reliable.
    GIT_TREE_STATE := $(if $(shell git status --porcelain),dirty,clean)
    # Use the tag if it exists, otherwise use the short commit hash.
    GIT_TAG      := $(if $(GIT_HEAD_TAG),$(GIT_HEAD_TAG),$(GIT_COMMIT))
endif

# If the tree is dirty, append a suffix to the version string.
ifneq ($(GIT_TREE_STATE), clean)
    GIT_VERSION := $(GIT_VERSION)-dirty
endif

# The import path for the shared version package within your framework.
# This is used by the linker to inject build information.
VERSION_PACKAGE_PATH := github.com/origadmin/toolkits/version

# Linker flags to inject version information into the binary
LDFLAGS := -X '$(VERSION_PACKAGE_PATH).Version=$(GIT_VERSION)' \
           -X '$(VERSION_PACKAGE_PATH).GitTag=$(GIT_TAG)' \
           -X '$(VERSION_PACKAGE_PATH).GitCommit=$(GIT_COMMIT)' \
           -X '$(VERSION_PACKAGE_PATH).GitBranch=$(GIT_BRANCH)' \
           -X '$(VERSION_PACKAGE_PATH).GitTreeState=$(GIT_TREE_STATE)' \
           -X '$(VERSION_PACKAGE_PATH).BuildDate=$(BUILD_DATE)'


# ---------------------------------------------------------------------------- #
#                  Version & Build Targets (LDFLAGS Usage Demo)                  #
# ---------------------------------------------------------------------------- #

.PHONY: version build release
.PHONY: $(addprefix build-, $(APPS)) $(addprefix release-, $(APPS))

version: ## ℹ️ Compile and run a version demo to show injected variables
	@echo "Compiling and running version demo..."
	@go run -ldflags="$(LDFLAGS)" ./cmd/version

build: $(addprefix build-, $(APPS)) ## 🔨 Build all application binaries with version info
release: $(addprefix release-, $(APPS)) ## 🚀 Create new releases for all applications

# Pattern rule for building a single application
# This injects the LDFLAGS into the final binary.
build-%:
	@echo "--> Building application: $*"
	@go build -ldflags="$(LDFLAGS)" -o ./bin/$* ./cmd/$*

# Pattern rule for releasing a single application
release-%:
	@echo "--> Releasing application: $*"
	@goreleaser release --clean -f ./cmd/$*/.goreleaser.yaml


# ---------------------------------------------------------------------------- #
#                                     Help                                     #
# ---------------------------------------------------------------------------- #

.PHONY: help
help: ## ✨ Show this help message
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Application-specific targets (from APPS variable):"
	@$(foreach app,$(APPS),printf "  \033[36m%-20s\033[0m %s\n", "build-$(app)", "Build the $(app) application";)
	@$(foreach app,$(APPS),printf "  \033[36m%-20s\033[0m %s\n", "release-$(app)", "Release the $(app) application";)

# Note: To use the 'lint' target, you need to install golangci-lint first.
# The 'make init' target can help with this.
